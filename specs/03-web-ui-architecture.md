# Web UI Architecture

## Overview

The web UI provides an intuitive interface for managing sing-box route rules using HTMX for dynamic interactions and Go templates for server-side rendering.

## Technology Choices

### HTMX
- **Why**: Achieve SPA-like experience without complex JavaScript frameworks
- **Benefits**:
  - Server-side rendering with dynamic updates
  - Reduced client-side complexity
  - Progressive enhancement
  - Better SEO and accessibility

### Go Templates
- **Why**: Type-safe, integrated with Go
- **Benefits**:
  - No additional templating language to learn
  - Compile-time safety
  - Native Go integration

### CSS Framework (TBD)
Options to consider:
- **Tailwind CSS**: Utility-first, flexible
- **Pure CSS**: Minimal, clean
- **Custom CSS**: Maximum control

## Application Structure

### HTTP Server

```go
internal/handlers/
├── server.go           # Server setup and routing
├── routes.go           # Route definitions
├── rule_handlers.go    # Rule CRUD handlers
├── ui_handlers.go      # Page rendering handlers
└── middleware.go       # Common middleware
```

### Handler Patterns

#### Page Handlers (Full HTML)
Return complete HTML pages for initial loads:
```go
func (h *Handler) RulesPage(w http.ResponseWriter, r *http.Request) {
    // GET /rules
    // Returns full HTML page
}
```

#### Partial Handlers (HTMX)
Return HTML fragments for dynamic updates:
```go
func (h *Handler) RuleForm(w http.ResponseWriter, r *http.Request) {
    // GET /rules/form?type=domain
    // Returns form HTML fragment
}

func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
    // POST /rules
    // Returns updated rules list fragment
}
```

## Page Structure

### Main Layout
```html
<!DOCTYPE html>
<html>
<head>
    <title>Sing-Box Config Manager</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <nav><!-- Navigation --></nav>
    <main>
        {{template "content" .}}
    </main>
</body>
</html>
```

### Rules Management Page

#### Layout
```
+--------------------------------------------------+
|  Sing-Box Config Manager                    [?] |
+--------------------------------------------------+
| Route Rules                                      |
+--------------------------------------------------+
| [+ Add Rule]  [Import]  [Export]                |
+--------------------------------------------------+
|                                                  |
| Rule 1: Domain Rule                    [Edit][X]|
| └─ Matches: google.com, facebook.com            |
| └─ Action: proxy                                |
|                                                  |
| Rule 2: GeoIP Rule                     [Edit][X]|
| └─ Country: CN                                  |
| └─ Action: direct                               |
|                                                  |
+--------------------------------------------------+
```

## HTMX Interactions

### Adding a Rule

1. **Click "+ Add Rule"**
   ```html
   <button hx-get="/rules/form"
           hx-target="#rule-modal"
           hx-swap="innerHTML">
       + Add Rule
   </button>
   ```

2. **Server returns form modal**
   ```html
   <div class="modal">
       <h2>Add Rule</h2>
       <select name="type"
               hx-get="/rules/form"
               hx-target="#form-content"
               hx-include="this">
           <option value="domain">Domain Rule</option>
           <option value="geoip">GeoIP Rule</option>
           <!-- ... -->
       </select>
       <div id="form-content">
           <!-- Type-specific form fields -->
       </div>
   </div>
   ```

3. **Submit form**
   ```html
   <form hx-post="/rules"
         hx-target="#rules-list"
         hx-swap="innerHTML">
       <!-- Form fields -->
       <button type="submit">Create Rule</button>
   </form>
   ```

4. **Server returns updated rules list**

### Editing a Rule

```html
<button hx-get="/rules/123/form"
        hx-target="#rule-modal"
        hx-swap="innerHTML">
    Edit
</button>
```

### Deleting a Rule

```html
<button hx-delete="/rules/123"
        hx-target="#rule-123"
        hx-swap="outerHTML"
        hx-confirm="Delete this rule?">
    Delete
</button>
```

### Real-time Validation

```html
<input name="domain"
       hx-post="/validate/domain"
       hx-trigger="keyup changed delay:500ms"
       hx-target="#domain-error">
```

## Template Organization

```
web/templates/
├── base.html              # Base layout
├── pages/
│   ├── index.html         # Home page
│   ├── rules.html         # Rules management page
│   └── settings.html      # Settings page
├── components/
│   ├── nav.html           # Navigation bar
│   ├── rule-list.html     # Rules list component
│   ├── rule-item.html     # Single rule display
│   └── modals.html        # Modal templates
└── forms/
    ├── domain-rule.html   # Domain rule form
    ├── geoip-rule.html    # GeoIP rule form
    └── ...                # Other rule type forms
```

## Data Flow

### Request → Response Flow

```
User Action
    ↓
HTMX Request (with headers)
    ↓
Router matches path
    ↓
Handler extracts parameters
    ↓
Business logic (validate, save, etc.)
    ↓
Render template with data
    ↓
Return HTML fragment
    ↓
HTMX swaps content in DOM
```

### State Management

#### Server-Side State
- **Session**: User preferences, temporary data
- **Database/File**: Persistent configuration
- **Memory**: Current configuration being edited

#### Client-Side State
- **Minimal**: HTMX handles most state via DOM
- **Form state**: Preserved in HTML
- **Temporary UI**: CSS classes for animations, highlights

## Form Generation

### Dynamic Form Fields

Based on rule type, generate appropriate form fields:

```go
type FormField struct {
    Name        string
    Type        string  // text, select, checkbox, textarea
    Label       string
    Required    bool
    Placeholder string
    Options     []string  // For select fields
    Validation  string
}

func GenerateFormFields(ruleType string) []FormField {
    // Return fields based on type
}
```

### Example: Domain Rule Form

```html
<form hx-post="/rules" class="rule-form">
    <input type="hidden" name="type" value="domain">

    <div class="field">
        <label for="domain">Domain</label>
        <input type="text" name="domain"
               placeholder="example.com">
        <small>Exact domain match</small>
    </div>

    <div class="field">
        <label for="domain_suffix">Domain Suffix</label>
        <input type="text" name="domain_suffix"
               placeholder=".example.com">
        <small>Match domain and subdomains</small>
    </div>

    <div class="field">
        <label for="outbound">Outbound</label>
        <select name="outbound">
            <option value="direct">Direct</option>
            <option value="proxy">Proxy</option>
            <option value="block">Block</option>
        </select>
    </div>

    <button type="submit">Create Rule</button>
</form>
```

## API Endpoints

### Page Endpoints
- `GET /` - Home page
- `GET /rules` - Rules management page
- `GET /export` - Export configuration page

### API Endpoints (HTMX)
- `GET /rules/form?type=X` - Get rule form for type X
- `GET /rules/:id/form` - Get edit form for rule ID
- `POST /rules` - Create new rule
- `PUT /rules/:id` - Update rule
- `DELETE /rules/:id` - Delete rule
- `POST /validate/:field` - Validate field value

### Data Endpoints
- `GET /api/rules` - Get all rules as JSON
- `POST /api/import` - Import configuration
- `GET /api/export` - Export configuration as JSON

## Error Handling

### Validation Errors
Display inline with HTMX:
```html
<div class="field">
    <input name="domain" ...>
    <div id="domain-error" class="error">
        <!-- HTMX inserts error here -->
    </div>
</div>
```

### Server Errors
Return appropriate HTTP status with error message:
```go
if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    tmpl.Execute(w, ErrorData{Message: err.Error()})
    return
}
```

### Client Feedback
- **Success**: Green highlight, fade after 2s
- **Error**: Red border, error message
- **Loading**: Spinner or disabled state

## Styling Strategy

### CSS Organization
```css
/* Base styles */
:root { /* CSS variables */ }
body { /* Layout */ }

/* Components */
.rule-list { }
.rule-item { }
.modal { }

/* Forms */
.form { }
.field { }

/* Utilities */
.error { }
.success { }
.loading { }
```

### Responsive Design
- Mobile-first approach
- Breakpoints: 640px, 768px, 1024px
- Stack on mobile, side-by-side on desktop

## Progressive Enhancement

1. **Base**: Works without JavaScript
2. **Enhanced**: HTMX adds dynamic behavior
3. **Polished**: CSS transitions and animations

## Accessibility

- Semantic HTML
- ARIA labels where needed
- Keyboard navigation
- Screen reader friendly
- Color contrast compliance

## Performance

### Optimization Strategies
- Minimal JavaScript (only HTMX)
- Server-side rendering (fast initial load)
- Efficient templates (pre-compile)
- CSS minification
- Lazy loading for large lists

### Caching
- Static assets: aggressive caching
- Templates: compile once at startup
- API responses: short cache for validation
