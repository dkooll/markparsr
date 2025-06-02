# Virtual Network

This terraform module simplifies the process of creating and managing virtual network resources on azure with configurable options for network topology, subnets, security groups, and more to ensure a secure and efficient environment for resource communication in the cloud.

<!-- BEGIN_TF_DOCS -->
## Requirements

The following requirements are needed by this module:

- <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) (>= 1.9.0)

- <a name="requirement_azurerm"></a> [azurerm](#requirement\_azurerm) (~> 4.0)

## Providers

The following providers are used by this module:

- <a name="provider_azurerm"></a> [azurerm](#provider\_azurerm) (~> 4.0)

## Resources

The following resources are used by this module:

- [azurerm_network_security_group.nsg](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/network_security_group) (resource)
- [azurerm_network_security_rule.rules](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/network_security_rule) (resource)
- [azurerm_route.routes](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/route) (resource)
- [azurerm_route_table.rt](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/route_table) (resource)
- [azurerm_subnet.subnets](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/subnet) (resource)
- [azurerm_subnet_network_security_group_association.nsg_as](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/subnet_network_security_group_association) (resource)
- [azurerm_subnet_route_table_association.rt_as](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/subnet_route_table_association) (resource)
- [azurerm_virtual_network.vnet](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network) (resource)
- [azurerm_virtual_network_dns_servers.dns](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network_dns_servers) (resource)
- [azurerm_virtual_network.existing](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/virtual_network) (data source)

## Required Inputs

The following input variables are required:

### <a name="input_vnet"></a> [vnet](#input\_vnet)

Description: Contains all virtual network configuration

Type:

```hcl
object({
    name                           = string
    address_space                  = optional(list(string))
    resource_group_name            = optional(string)
    location                       = optional(string)
    use_existing_vnet              = optional(bool, false)
    edge_zone                      = optional(string)
    bgp_community                  = optional(string)
    flow_timeout_in_minutes        = optional(number)
    private_endpoint_vnet_policies = optional(string)
    dns_servers                    = optional(list(string), [])
    tags                           = optional(map(string))
    ddos_protection_plan = optional(object({
      id     = string
      enable = optional(bool, true)
    }))
    encryption = optional(object({
      enforcement = string
    }))
    subnets = optional(map(object({
      name                                          = optional(string)
      address_prefixes                              = list(string)
      service_endpoints                             = optional(list(string), [])
      private_link_service_network_policies_enabled = optional(bool, false)
      private_endpoint_network_policies             = optional(string, "Disabled")
      service_endpoint_policy_ids                   = optional(list(string), [])
      default_outbound_access_enabled               = optional(bool, null)
      delegations = optional(map(object({
        name    = string
        actions = optional(list(string), [])
      })), {})
      network_security_group = optional(object({
        name = optional(string)
        rules = optional(map(object({
          name                                       = optional(string)
          priority                                   = number
          direction                                  = string
          access                                     = string
          protocol                                   = string
          source_port_range                          = optional(string)
          source_port_ranges                         = optional(list(string))
          destination_port_range                     = optional(string)
          destination_port_ranges                    = optional(list(string))
          source_address_prefix                      = optional(string)
          source_address_prefixes                    = optional(list(string))
          destination_address_prefix                 = optional(string)
          destination_address_prefixes               = optional(list(string))
          description                                = optional(string)
          source_application_security_group_ids      = optional(list(string), [])
          destination_application_security_group_ids = optional(list(string), [])
        })), {})
      }))
      route_table = optional(object({
        name                          = optional(string)
        bgp_route_propagation_enabled = optional(bool, true)
        routes = optional(map(object({
          name                   = optional(string)
          address_prefix         = string
          next_hop_type          = string
          next_hop_in_ip_address = optional(string, null)
        })), {})
      }))
      shared = optional(object({
        network_security_group = optional(string)
        route_table            = optional(string)
      }), {})
    })), {})
    network_security_groups = optional(map(object({
      name = optional(string)
      rules = optional(map(object({
        name                                       = optional(string)
        priority                                   = number
        direction                                  = string
        access                                     = string
        protocol                                   = string
        source_port_range                          = optional(string)
        source_port_ranges                         = optional(list(string), null)
        destination_port_range                     = optional(string, null)
        destination_port_ranges                    = optional(list(string), null)
        source_address_prefix                      = optional(string, null)
        source_address_prefixes                    = optional(list(string), null)
        destination_address_prefix                 = optional(string, null)
        destination_address_prefixes               = optional(list(string), null)
        description                                = optional(string, null)
        source_application_security_group_ids      = optional(list(string), [])
        destination_application_security_group_ids = optional(list(string), [])
      })), {})
    })), {})
    route_tables = optional(map(object({
      name                          = optional(string)
      bgp_route_propagation_enabled = optional(bool, true)
      routes = optional(map(object({
        name                   = optional(string)
        address_prefix         = string
        next_hop_type          = string
        next_hop_in_ip_address = optional(string, null)
      })), {})
    })), {})
  })
```

## Optional Inputs

The following input variables are optional (have default values):

### <a name="input_location"></a> [location](#input\_location)

Description: default azure region to be used.

Type: `string`

Default: `null`

### <a name="input_naming"></a> [naming](#input\_naming)

Description: Used for naming purposes

Type: `map(string)`

Default: `null`

### <a name="input_resource_group_name"></a> [resource\_group\_name](#input\_resource\_group\_name)

Description: default resource group to be used.

Type: `string`

Default: `null`

### <a name="input_tags"></a> [tags](#input\_tags)

Description: tags to be added to the resources

Type: `map(string)`

Default: `{}`

### <a name="input_use_existing_vnet"></a> [use\_existing\_vnet](#input\_use\_existing\_vnet)

Description: Whether to use existing VNet for all vnets

Type: `bool`

Default: `false`

## Outputs

The following outputs are exported:

### <a name="output_network_security_group"></a> [network\_security\_group](#output\_network\_security\_group)

Description: contains network security group configuration

### <a name="output_subnets"></a> [subnets](#output\_subnets)

Description: contains subnet configuration

### <a name="output_vnet"></a> [vnet](#output\_vnet)

Description: contains virtual network configuration
<!-- END_TF_DOCS -->

## Goals

For more information, please see our [goals and non-goals](./GOALS.md).

## Testing

For more information, please see our testing [guidelines](./TESTING.md)

## Notes

This is an experimental module for private use.

A naming convention is developed using regular expressions to ensure correct abbreviations, with flexibility for multiple prefixes and suffixes.

Full usage examples and integrations with dependency modules are in the examples directory.

To update the module's documentation run `make doc`
