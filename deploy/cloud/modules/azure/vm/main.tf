resource "random_id" "id" {
  byte_length = 4
}

resource "azurerm_resource_group" "resource_group" {
  name     = "${local.deploy_name}-res-group"
  location = var.location
}

locals {
  vm_private_key_file = "${path.module}/azure-vm-${random_id.id.hex}.pem"
  vm_username         = "azureuser"
  deploy_name         = "${var.deployment_name}-${random_id.id.hex}"
  tags = merge({
    name = var.deployment_name
  }, var.specific_tags)
}

resource "tls_private_key" "azure_vm_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "local_file" "ssh_private_key" {
  filename        = local.vm_private_key_file
  content         = tls_private_key.azure_vm_key.private_key_pem
  file_permission = 0400
}

#=== Network Configuration ===

resource "azurerm_virtual_network" "vm_virtual_network" {
  name                = "${local.deploy_name}-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
}

resource "azurerm_subnet" "vm_subnet" {
  name                 = "${local.deploy_name}-internal"
  resource_group_name  = azurerm_resource_group.resource_group.name
  virtual_network_name = azurerm_virtual_network.vm_virtual_network.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_public_ip" "pip" {
  name                    = "${local.deploy_name}-pip"
  location                = azurerm_resource_group.resource_group.location
  resource_group_name     = azurerm_resource_group.resource_group.name
  allocation_method       = "Dynamic"
  idle_timeout_in_minutes = 30

  tags = local.tags
}

resource "azurerm_network_interface" "vm_nic" {
  name                = "${local.deploy_name}-nic"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.vm_subnet.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.pip.id
  }
}

resource "azurerm_network_security_group" "nsg" {
  name                = "${local.deploy_name}-nsg"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name

  security_rule {
    name                       = "AllowSSHInbound"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "AllowAllOutbound"
    priority                   = 200
    direction                  = "Outbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "*"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  tags = local.tags
}

resource "azurerm_network_interface_security_group_association" "azure_vm_nsg_association" {
  network_interface_id      = azurerm_network_interface.vm_nic.id
  network_security_group_id = azurerm_network_security_group.nsg.id
}

data "azurerm_public_ip" "data-pip" {
  name                = azurerm_public_ip.pip.name
  resource_group_name = azurerm_linux_virtual_machine.linux_vm.resource_group_name
}

# ==========================================

resource "azurerm_linux_virtual_machine" "linux_vm" {
  name                = "${local.deploy_name}-vm"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  size                = var.size
  admin_username      = local.vm_username
  network_interface_ids = [
    azurerm_network_interface.vm_nic.id
  ]

  admin_ssh_key {
    username   = local.vm_username
    public_key = tls_private_key.azure_vm_key.public_key_openssh
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts"
    version   = "latest"
  }
  tags = local.tags
}