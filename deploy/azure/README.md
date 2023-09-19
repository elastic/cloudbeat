## Azure deployment

This is a work in progress.

Deploy the JSON template at https://portal.azure.com/#create/Microsoft.Template.

To be able to ssh into the vm, you need to change the password before deploying:

```diff
diff --git a/deploy/azure/azureARMTemplate.json b/deploy/azure/azureARMTemplate.json
index 41defb01..f97234e3 100644
--- a/deploy/azure/azureARMTemplate.json
+++ b/deploy/azure/azureARMTemplate.json
@@ -58,7 +58,7 @@
                 "osProfile": {
                     "computerName": "cloudbeat",
                     "adminUsername": "cloudbeatVM",
-                    "adminPassword": "[guid('')]",
+                    "adminPassword": "My-password123!",
                     "linuxConfiguration": {
                         "disablePasswordAuthentication": false
                     }
```
