# Hướng Dẫn Sử Dụng Terraform Provider ViettelIDC

## Cài Đặt Provider

```hcl
terraform {
  required_providers {
    viettelidc = {
      source  = "viettelidc.com.vn/iac/viettelidc"
      version = "0.1.0"
    }
  }
}

provider "viettelidc" {
  base_url    = "https://iac.viettelidc.com.vn"  # URL API Gateway (mặc định)
  customer_id = "<customer_id>"
  token       = "<bearer_token>"           # sensitive
  vpc_id      = "<default_vpc_id>"         # VPC mặc định cho toàn bộ resource
}
```

---

## Danh Sách Resource

| Resource | Mô tả | Async (polling) |
|---|---|---|
| `viettelidc_vpc` | Tạo VPC | ✅ 2 phút |
| `viettelidc_subnet` | Tạo subnet trong VPC | ✅ |
| `viettelidc_nat_gateway` | NAT Gateway cho subnet | ✅ 5 phút |
| `viettelidc_load_balancer` | Load Balancer (Application/Network) | ✅ 10 phút |
| `viettelidc_backup_plan` | Kế hoạch sao lưu tự động | ✅ |
| `viettelidc_security_group` | Security Group | ✅ |
| `viettelidc_security_group_rule` | Rule cho Security Group | ✅ |
| `viettelidc_network_interface` | Network Interface (NIC) | ✅ |
| `viettelidc_network_interface_attachment` | Gắn NIC vào instance | đồng bộ |
| `viettelidc_instance` | Máy ảo (VM) | ✅ |
| `viettelidc_volume` | Block Storage Volume | ✅ |
| `viettelidc_volume_attachment` | Gắn volume vào instance | đồng bộ |
| `viettelidc_floating_ip` | IP public gắn vào instance | đồng bộ |
| `viettelidc_key_pair` | SSH Key Pair | đồng bộ |
| `viettelidc_route_table` | Route Table | đồng bộ |
| `viettelidc_route_table_association` | Gắn route table vào subnet | đồng bộ |
| `viettelidc_certificate` | TLS Certificate (dùng cho LB) | ✅ |

> **✅ Async**: Terraform sẽ chờ đến khi resource sẵn sàng. Nếu timeout, `apply` báo lỗi — kiểm tra trạng thái trên portal.

---

## Ví Dụ Tối Giản: VPC → Subnet → Instance

```hcl
resource "viettelidc_vpc" "main" {
  name = "my-vpc"
  cidr = "10.0.0.0/16"
}

resource "viettelidc_subnet" "app" {
  name            = "app-subnet"
  network_address = "10.0.1.0/24"
  is_public_zone  = false
  depends_on      = [viettelidc_vpc.main]
}

resource "viettelidc_security_group" "app" {
  name = "app-sg"
}

resource "viettelidc_instance" "app" {
  template_id        = var.template_id
  admin_pass         = var.admin_pass
  cpu                = 2
  memory             = 4096
  subnet_id          = viettelidc_subnet.app.id
  availability_zone  = "AZ1"
  security_group_ids = [viettelidc_security_group.app.id]

  depends_on = [viettelidc_subnet.app, viettelidc_security_group.app]
}
```

---

## Import Resource Vào State

```bash
# Cú pháp chung
terraform import <resource_type>.<name> <id>

# Ví dụ
terraform import viettelidc_subnet.app 12345
terraform import viettelidc_instance.app 67890
```

---

## Lưu Ý Quan Trọng

| Vấn đề | Chi tiết |
|---|---|
| `route_table` không bị xoá | Hệ thống không có API xoá route table — `destroy` chỉ xoá khỏi state |
| Async timeout | Nếu tạo resource quá thời gian polling → Terraform báo lỗi, kiểm tra trên portal |
| `key_pair` download | `download_url` chỉ hợp lệ **một lần** ngay sau khi tạo |
| `token` sensitive | Không commit token vào git — dùng biến môi trường hoặc Vault |

---

## Biến Môi Trường

```bash
export VIETTELIDC_BASE_URL="https://iac.viettelidc.com.vn"
export VIETTELIDC_CUSTOMER_ID="12345"
export TF_VAR_token="eyJ..."
export TF_VAR_vpc_id="99"
```

---

## Demo Đầy Đủ

Xem [`examples/iac/demo/main.tf`](../../../examples/iac/demo/main.tf) — kịch bản hoàn chỉnh:
subnet → route table → security group → keypair → NIC → instance → volume → floating IP.
