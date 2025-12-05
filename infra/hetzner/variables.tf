variable "server_id" {
  description = "Existing Hetzner server ID"
  type        = number
}

variable "my_home_ip" {
  description = "Public IPv4 of your current location, without /32"
  type        = string
}
