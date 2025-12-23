data "google_client_config" "default" {}

resource "terraform_data" "validate_startup" {
  count = var.enabled ? 1 : 0

  provisioner "local-exec" {
    command = "bash ${path.module}/validate_startup.sh '${var.project_id}' '${var.zone}' '${var.instance_name}' '${var.timeout}'"
    environment = {
      GCP_ACCESS_TOKEN = data.google_client_config.default.access_token
    }
  }

  triggers_replace = {
    instance_id = var.instance_id
  }
}
