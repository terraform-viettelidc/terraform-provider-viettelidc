##############################################################################
# ViettelIDC IaC – Full Demo Scenario
#
# Kịch bản: Tạo đầy đủ hạ tầng từ đầu:
#   1. Subnet
#   2. Route Table + gắn subnet
#   3. Security Group + 2 inbound rules (SSH + HTTP)
#   4. Key Pair (SSH)
#   5. Network Interface
#   6. Instance (gắn SG + NIC)
#   7. Volume + gắn vào instance
#   8. Floating IP + gắn vào instance
#
# Sử dụng:
#   terraform init
#   terraform apply
#   terraform destroy  (Lưu ý: route_table KHÔNG bị xoá trên cloud)
##############################################################################

terraform {
  required_providers {
    viettelidc = {
      source  = "viettelidc.com.vn/iac/viettelidc"
      version = "0.1.0"
    }
  }
}

provider "viettelidc" {
  base_url    = var.base_url
  customer_id = var.customer_id
  token       = var.token
  vpc_id      = var.vpc_id
}

##############################################################################
# Variables
##############################################################################

variable "base_url" {
  type        = string
  description = "URL của API Gateway (ví dụ: https://iac.viettelidc.com.vn)"
}

variable "customer_id" {
  type        = string
  description = "Customer ID của tenant"
}

variable "token" {
  type        = string
  sensitive   = true
  description = "Bearer token xác thực CSA"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID dùng cho toàn bộ kịch bản"
}

variable "availability_zone" {
  type        = string
  default     = "AZ1"
  description = "Availability Zone"
}

variable "admin_pass" {
  type        = string
  sensitive   = true
  description = "Mật khẩu admin của instance"
}

variable "template_id" {
  type        = string
  description = "Template ID (flavor) dùng để tạo instance"
}

##############################################################################
# 1. Subnet
##############################################################################

resource "viettelidc_ovpc_subnet" "demo" {
  name            = "demo-subnet"
  network_address = "192.168.100.0/24"
  is_public_zone  = false
  description     = "Subnet demo"
}

##############################################################################
# 2. Route Table + Association
##############################################################################

resource "viettelidc_ovpc_route_table" "demo" {
  name = "demo-rt"

  # vpc_id kế thừa từ provider nếu không set
}

resource "viettelidc_ovpc_route_table_association" "demo" {
  route_table_id = viettelidc_ovpc_route_table.demo.id
  subnet_id      = viettelidc_ovpc_subnet.demo.id

  depends_on = [viettelidc_ovpc_subnet.demo, viettelidc_ovpc_route_table.demo]
}

##############################################################################
# 3. Security Group + Rules
##############################################################################

resource "viettelidc_ovpc_security_group" "demo" {
  name        = "demo-sg"
  description = "Security Group dùng cho demo"
}

# Inbound SSH (port 22)
resource "viettelidc_ovpc_security_group_rule" "ssh_in" {
  security_group_id = viettelidc_ovpc_security_group.demo.id
  direction         = "Inbound"
  rule_type         = "Custom"
  protocol_name     = "TCP"
  port              = "22"
  source_ip         = "0.0.0.0/0"
  action            = "Allow"
}

# Inbound HTTP (port 80)
resource "viettelidc_ovpc_security_group_rule" "http_in" {
  security_group_id = viettelidc_ovpc_security_group.demo.id
  direction         = "Inbound"
  rule_type         = "Custom"
  protocol_name     = "TCP"
  port              = "80"
  source_ip         = "0.0.0.0/0"
  action            = "Allow"
}

##############################################################################
# 4. Key Pair
##############################################################################

resource "viettelidc_ovpc_key_pair" "demo" {
  key_name = "demo-keypair"
}

# URL tải private key (một lần duy nhất sau khi tạo)
output "key_pair_download_url" {
  value     = viettelidc_ovpc_key_pair.demo.download_url
  sensitive = true
}

##############################################################################
# 5. Network Interface
##############################################################################

resource "viettelidc_ovpc_network_interface" "demo" {
  name           = "demo-nic"
  subnet_id      = viettelidc_ovpc_subnet.demo.id
  ip_assign_type = "DYNAMIC"
  description    = "NIC demo"
}

##############################################################################
# 6. Instance
##############################################################################

resource "viettelidc_ovpc_instance" "demo" {
  template_id       = var.template_id
  admin_pass        = var.admin_pass
  cpu               = 2
  memory            = 4096
  subnet_id         = viettelidc_ovpc_subnet.demo.id
  availability_zone = var.availability_zone

  security_group_ids = [
    viettelidc_ovpc_security_group.demo.id,
  ]

  depends_on = [
    viettelidc_ovpc_subnet.demo,
    viettelidc_ovpc_security_group.demo,
    viettelidc_ovpc_security_group_rule.ssh_in,
    viettelidc_ovpc_security_group_rule.http_in,
  ]
}

# Gắn NIC thêm vào instance
resource "viettelidc_ovpc_network_interface_attachment" "demo" {
  network_interface_id = viettelidc_ovpc_network_interface.demo.id
  instance_id          = viettelidc_ovpc_instance.demo.id
}

##############################################################################
# 7. Volume + Attachment
##############################################################################

resource "viettelidc_ovpc_volume" "demo" {
  name        = "demo-vol"
  size        = 50
  volume_type = "SSD"
}

resource "viettelidc_ovpc_volume_attachment" "demo" {
  instance_id = viettelidc_ovpc_instance.demo.id
  volume_id   = viettelidc_ovpc_volume.demo.id
}

##############################################################################
# 8. Floating IP + Associate với instance
##############################################################################

resource "viettelidc_ovpc_floating_ip" "demo" {
  instance_id          = viettelidc_ovpc_instance.demo.id
  network_interface_id = viettelidc_ovpc_network_interface.demo.id
}

##############################################################################
# Outputs
##############################################################################

output "subnet_id" {
  value = viettelidc_ovpc_subnet.demo.id
}

output "route_table_id" {
  value = viettelidc_ovpc_route_table.demo.id
}

output "security_group_id" {
  value = viettelidc_ovpc_security_group.demo.id
}

output "instance_id" {
  value = viettelidc_ovpc_instance.demo.id
}

output "instance_ip_address" {
  value = viettelidc_ovpc_instance.demo.ip_address
}

output "instance_status" {
  value = viettelidc_ovpc_instance.demo.status
}

output "volume_id" {
  value = viettelidc_ovpc_volume.demo.id
}

output "floating_ip" {
  value = viettelidc_ovpc_floating_ip.demo.public_ip
}
