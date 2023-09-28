## Azure deployment

This is a work in progress.

Deploy the JSON template at https://portal.azure.com/#create/Microsoft.Template.

To be able to ssh into the vm, you need to change the password before deploying and also to remove the installation
script as it resets the password and disables ssh. You'll need to install the agent manually after ssh-ing into the
machine.

```diff
diff --git a/deploy/azure/azureARMTemplate.json b/deploy/azure/azureARMTemplate.json
index 2a3365c8..f17ed213 100644
--- a/deploy/azure/azureARMTemplate.json
+++ b/deploy/azure/azureARMTemplate.json
@@ -65,7 +65,7 @@
                 "osProfile": {
                     "computerName": "cloudbeatVM",
                     "adminUsername": "cloudbeat",
-                    "adminPassword": "[concat('Salt123@', guid(parameters('Seed')))]",
+                    "adminPassword": "My-password123!",
                     "linuxConfiguration": {
                         "disablePasswordAuthentication": false
                     }
@@ -153,26 +153,6 @@
                 "publicIPAllocationMethod": "Dynamic"
             }
         },
-        {
-            "type": "Microsoft.Compute/virtualMachines/extensions",
-            "apiVersion": "2021-04-01",
-            "name": "cloudbeatVM/customScriptExtension",
-            "location": "[resourceGroup().location]",
-            "dependsOn": [
-                "[resourceId('Microsoft.Compute/virtualMachines', 'cloudbeatVM')]"
-            ],
-            "properties": {
-                "publisher": "Microsoft.Azure.Extensions",
-                "type": "CustomScript",
-                "typeHandlerVersion": "2.1",
-                "settings": {
-                    "fileUris": [
-                        "https://raw.githubusercontent.com/elastic/cloudbeat/main/deploy/azure/install-agent.sh"
-                    ],
-                    "commandToExecute": "[concat('bash install-agent.sh ', parameters('ElasticAgentVersion'), ' ', parameters('ElasticArtifactServer'), ' ', parameters('FleetUrl'), ' ', parameters('EnrollmentToken'))]"
-                }
-            }
-        },
         {
             "type": "Microsoft.Authorization/roleAssignments",
             "apiVersion": "2022-04-01",
```
