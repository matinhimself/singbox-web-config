# Sing-Box Route Rules Reference

## Overview

Sing-box route rules determine how network traffic is routed based on various matching criteria. Each rule can match traffic based on multiple conditions and route it to a specific outbound.

## Rule Structure

### Basic Format

```json
{
  "route": {
    "rules": [
      {
        "type": "default",
        "inbound": ["..."],
        "domain": ["..."],
        "domain_suffix": ["..."],
        "geoip": ["..."],
        "outbound": "proxy"
      }
    ]
  }
}
```

## Rule Types

Based on `github.com/SagerNet/sing-box/route/rule`, the following types exist:

### 1. Default Rule
The standard rule type with multiple matching options.

**Match Options:**
- `inbound`: Match by inbound tag
- `protocol`: Match by protocol (tcp, udp, etc.)
- `network`: Match by network type
- `domain`: Exact domain match
- `domain_suffix`: Domain suffix match
- `domain_keyword`: Domain contains keyword
- `domain_regex`: Domain regex match
- `geosite`: Match by geosite database
- `source_geoip`: Match by source IP geography
- `geoip`: Match by destination IP geography
- `source_ip_cidr`: Match by source IP CIDR
- `ip_cidr`: Match by destination IP CIDR
- `source_port`: Match by source port
- `source_port_range`: Match by source port range
- `port`: Match by destination port
- `port_range`: Match by destination port range
- `process_name`: Match by process name
- `process_path`: Match by process path
- `user`: Match by user
- `user_id`: Match by user ID

**Action:**
- `outbound`: Target outbound tag

### 2. Logical Rules

#### AND Rule
All sub-rules must match:
```json
{
  "type": "logical",
  "mode": "and",
  "rules": [
    { "domain": ["example.com"] },
    { "port": [80, 443] }
  ],
  "outbound": "proxy"
}
```

#### OR Rule
Any sub-rule must match:
```json
{
  "type": "logical",
  "mode": "or",
  "rules": [
    { "domain": ["example.com"] },
    { "domain": ["example.org"] }
  ],
  "outbound": "proxy"
}
```

## Common Use Cases

### 1. Bypass Chinese Websites
```json
{
  "geosite": ["cn"],
  "outbound": "direct"
}
```

### 2. Route by Domain
```json
{
  "domain": ["google.com"],
  "domain_suffix": [".google.com"],
  "outbound": "proxy"
}
```

### 3. Block Ads
```json
{
  "geosite": ["category-ads-all"],
  "outbound": "block"
}
```

### 4. Route Private IPs
```json
{
  "ip_cidr": [
    "192.168.0.0/16",
    "10.0.0.0/8"
  ],
  "outbound": "direct"
}
```

### 5. Route Specific Ports
```json
{
  "port": [80, 443],
  "protocol": ["tcp"],
  "outbound": "proxy"
}
```

## Rule Priority

Rules are matched in order from top to bottom. First match wins.

**Best Practice:**
1. Specific rules first (exact domains)
2. Broad rules after (CIDR blocks)
3. Default/catch-all rule last

## Field Types Reference

### String Arrays
Fields that accept multiple string values:
- `inbound`
- `domain`
- `domain_suffix`
- `domain_keyword`
- `domain_regex`
- `geosite`
- `geoip`
- `source_geoip`
- `ip_cidr`
- `source_ip_cidr`

### Integer Arrays
Fields that accept multiple integer values:
- `port`
- `source_port`

### String
Single string fields:
- `outbound`
- `protocol`
- `network`
- `process_name`
- `process_path`

### Port Ranges
Special format:
- `port_range`: ["1000:2000", "3000:4000"]
- `source_port_range`: ["1000:2000"]

## Validation Rules

### Domain Validation
- Must be valid domain format
- Can include wildcards (*.example.com)
- No spaces or special characters except dots, hyphens

### IP/CIDR Validation
- Must be valid IPv4 or IPv6
- CIDR notation required for ranges (e.g., 192.168.1.0/24)

### Port Validation
- Must be 1-65535
- Port ranges: "start:end" where start < end

### Regex Validation
- Must be valid Go regex pattern
- Will be compiled and validated

## Data Sources

### GeoIP Databases
- Based on MaxMind GeoIP2
- Updated regularly
- Country codes: ISO 3166-1 alpha-2 (CN, US, etc.)

### GeoSite Databases
- Based on v2ray/domain-list-community
- Categories:
  - `cn`: Chinese websites
  - `geolocation-!cn`: Non-Chinese websites
  - `category-ads-all`: Ad domains
  - `google`: Google services
  - `facebook`: Facebook services
  - Many more...

## Generator Considerations

When generating types from sing-box source:

### 1. Optional Fields
Most fields are optional. Use `omitempty` in JSON tags:
```go
Domain []string `json:"domain,omitempty"`
```

### 2. Type Mapping
- Go `string` → UI text input
- Go `[]string` → UI multiple input or textarea
- Go `uint16` → UI number input (for ports)
- Go `bool` → UI checkbox

### 3. Validation Tags
Extract and preserve validation rules:
```go
Port []uint16 `json:"port,omitempty" validate:"dive,min=1,max=65535"`
```

### 4. Documentation
Preserve comments for UI tooltips:
```go
// Domain specifies domains to match exactly
Domain []string `json:"domain,omitempty"`
```

## UI Representation Strategy

### Rule Type Selection
Dropdown with all available rule types.

### Dynamic Form Fields
Based on selected rule type, show relevant fields.

### Multi-value Inputs
For string arrays:
- Tags input (add/remove)
- Textarea (one per line)
- Multiple text inputs

### Validation
Real-time validation as user types:
- Domain format
- IP/CIDR format
- Port ranges
- Regex compilation

## Configuration Export

### JSON Generation
Convert UI form data to valid sing-box JSON:
```go
func (r *Rule) ToJSON() ([]byte, error) {
    return json.MarshalIndent(r, "", "  ")
}
```

### Configuration File
Generate complete sing-box config:
```json
{
  "route": {
    "rules": [
      // Generated rules here
    ],
    "final": "proxy"
  }
}
```

## Testing Strategy

### Valid Configurations
Test with known-good sing-box configs.

### Invalid Configurations
Test validation catches errors:
- Invalid domains
- Invalid IP addresses
- Out-of-range ports
- Bad regex patterns

### Edge Cases
- Empty rules
- Rules with all optional fields
- Complex logical rules
- Very long arrays

## Future Enhancements

1. **Rule Templates**: Pre-configured common rules
2. **Rule Testing**: Test rules against sample traffic
3. **Rule Suggestions**: AI-powered rule suggestions
4. **Rule Import**: Parse existing configs
5. **Rule Validation**: Validate against sing-box binary
6. **GeoIP Preview**: Show which IPs match a rule
7. **Domain Testing**: Test if domain matches rule
