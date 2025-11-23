# Type Generation Notes

This document tracks types and fields that couldn't be fully generated or have limitations.

## Successfully Generated Types

### Config Structure
- ✅ `Config` - Main configuration type
- ✅ `LogOptions` - Log configuration
- ✅ `RawDNSOptions` - DNS configuration (16 types generated)
- ✅ `NTPOptions` - NTP configuration
- ✅ `RouteOptions` - Routing configuration (3 types generated)
- ✅ `ExperimentalOptions` - Experimental features (5 types generated)

### Rules
- ✅ Rule types (19 types generated including RawDefaultRule, RawLogicalRule, DNS rules, etc.)

### Inbounds/Outbounds
- ✅ Basic inbound/outbound options generated (4 inbound types, 5 outbound types)

## Types That Couldn't Be Fully Generated

### 1. Complex sing-box Custom Types

The following sing-box custom types were simplified to basic Go types or `interface{}`:

- `badoption.Listable[T]` → `[]T`
- `badoption.Prefixable` / `badPrefix` → `string`
- `badoption.Duration` → `uint32`
- `badjson.TypedMap[K,V]` → `map[string]interface{}`
- `badAddr` → `string`
- `badHTTPHeader` → `map[string]string`
- `FwMark` → `uint32`
- `UDPTimeoutCompat` → `uint32`
- `DomainResolveOptions` → `interface{}`
- `DebugOptions` → `interface{}`

**Reason**: These are custom sing-box types that use generics or complex type systems. Simplifying them to basic types maintains compatibility while losing some type safety.

### 2. Polymorphic Types

The following were converted to `interface{}` or `[]interface{}`:

- `Inbound` / `[]Inbound`
- `Outbound` / `[]Outbound`
- `Endpoint` / `[]Endpoint`
- `Service` / `[]Service`
- `RuleSet` / `[]RuleSet`
- `DNSServerOptions` / `[]DNSServerOptions`
- `Rule` / `[]Rule`
- `DNSRule` / `[]DNSRule`
- `HeadlessRule` / `[]HeadlessRule`

**Reason**: These are polymorphic types that can be one of many concrete implementations (e.g., inbound can be http, socks, mixed, etc.). The sing-box library uses custom unmarshalling logic to determine the correct type at runtime.

### 3. Missing or Deprecated Fields

- ❌ `route.rule_action` - This field was referenced in old handler code but doesn't exist in the current sing-box schema
- ℹ️ `log.disable_color` - Exists in sing-box but marked with `json:"-"` (internal field), not included in generated types

### 4. Interface Types Without Implementations

Some types are defined as interfaces in sing-box:

- `InboundOptionsRegistry`
- `OutboundOptionsRegistry`
- `DNSTransportOptionsRegistry`
- `ListenOptionsWrapper`
- `DialerOptionsWrapper`
- `ServerOptionsWrapper`

These were generated as empty structs since they're interfaces that don't have JSON representation.

## Recommendations

### For Full Type Safety

If you need full type safety for inbounds/outbounds, you would need to:

1. Define concrete types for each inbound/outbound type (e.g., `HTTPInbound`, `SocksInbound`, etc.)
2. Implement custom JSON unmarshalling logic similar to sing-box
3. Use type assertions or type switches when working with these types

### Current Approach

The current implementation uses `interface{}` for polymorphic types, which:
- ✅ Works with any valid sing-box configuration
- ✅ Easy to marshal/unmarshal JSON
- ⚠️ Loses compile-time type safety
- ⚠️ Requires runtime type assertions/checks

## Future Improvements

1. **Generate concrete inbound/outbound types**: Parse all inbound/outbound implementation files and generate specific types
2. **Custom unmarshalling**: Implement custom JSON unmarshalling for polymorphic types
3. **Validation helpers**: Add functions to validate configurations at runtime
4. **Builder patterns**: Add fluent API builders for common configurations
