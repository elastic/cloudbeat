## Azure deployment

This is a work in progress.

Deploy the JSON template at https://portal.azure.com/#create/Microsoft.Template.

To be able to ssh into the vm, apply this patch before deploying:

```diff
diff --git a/deploy/azure/azureARMTemplate.json b/deploy/azure/azureARMTemplate.json
index 6119c6ff..94528b1f 100644
--- a/deploy/azure/azureARMTemplate.json
+++ b/deploy/azure/azureARMTemplate.json
@@ -64,6 +64,8 @@
                 },
                 "osProfile": {
                     "computerName": "cloudbeatVM",
+                    "adminUsername": "<username>",
+                    "adminPassword": "<password here, needs lower case, upper case, numbers and special chars>"
                 },
                 "networkProfile": {
                     "networkInterfaces": [
```
