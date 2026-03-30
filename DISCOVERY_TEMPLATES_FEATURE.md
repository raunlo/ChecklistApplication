# Templates/Recipes Feature - Product Discovery Document

**Status:** MVP v1 Planning ✅ Complete
**Last Updated:** 2026-03-01
**Owner:** Rauno

---

## 1. Problem & Vision

### Current State
ChecklistApplication has evolved from a todo list into a flexible list management tool (todos, shopping lists, recipes, etc.). Users currently have to manually recreate common ingredient/item lists repeatedly.

### Desired State
Users can **save reusable ingredient templates** and quickly apply them to checklists with **smart duplicate detection**. When applying a template, the app shows which items already exist, and the user selects which ones to actually add.

**Killer Use Case:**
1. User makes Carbonara, saves ingredients as template: eggs, pasta, pancetta, pecorino, guanciale
2. Next time, opens Shopping checklist which already has: eggs, pasta
3. Applies Carbonara template
4. App shows: "You already have eggs & pasta. Which new items do you want to add?"
5. User unchecks eggs & pasta
6. Only pancetta, pecorino, guanciale are added

### Key Design Constraint
**App is mobile-first.** All UI must be compact and touch-friendly.

---

## 2. Core Decisions (User-Confirmed)

| Decision | Answer | Rationale |
|----------|--------|-----------|
| **Template Scope** | Items/Ingredients only | Templates are flat lists of ingredients/items (no nested structure). Carbonara = [eggs, pasta, pancetta, pecorino, guanciale]. Keeps it simple. |
| **Primary User Flows** | Both (in-checklist + templates page) | Users apply templates from within a checklist (most common) AND browse/manage templates from a dedicated page. |
| **Duplicate Handling** | Smart detection + user choice | When applying template, app detects items already in checklist and shows a picker. User selects which items to add. Prevents unwanted duplicates but allows intentional ones. |
| **Template Size** | Max 20 items | Safety limit for mobile performance. |
| **Instructions/Description** | Optional text field | Templates can have a description field (e.g., "Classic pasta carbonara recipe - prep time 20 mins"). Stored as text, shown in template preview. |
| **System Templates (MVP)** | Deferred to v2 | MVP focuses on user-created templates. Pre-built recipes come later. |
| **Approach** | Simple now, adapt later | Build MVP lean and iterate based on real usage patterns. |

---

## 3. User Stories & Acceptance Criteria

### Story 1: Create a Template from Items
**As a** user
**I want to** select multiple items from my checklist and save them as a reusable template
**So that** I can quickly add those items to other checklists later

**Acceptance Criteria:**
- [ ] User can long-press or menu an item to see "Save as template" option
- [ ] User is prompted to enter template name and optional description
- [ ] Multiple items can be selected at once (checkbox UI)
- [ ] Template is saved with all selected items
- [ ] Template appears in user's template library (Templates page)
- [ ] User sees confirmation after save

### Story 2: Apply Template with Smart Duplicates
**As a** user
**I want to** apply a template to my checklist, but skip items I already have
**So that** I don't end up with duplicate entries

**Acceptance Criteria:**
- [ ] User taps "Add template" button in checklist (floating action button)
- [ ] Searchable template picker appears (bottom sheet, Apple style)
- [ ] User selects a template
- [ ] App shows a preview: template items + detection of duplicates
- [ ] Items already in checklist are marked "Already have"
- [ ] User can check/uncheck which items to add
- [ ] Only selected items are added to checklist
- [ ] Change is reflected in real-time for other connected clients (SSE)

### Story 3: Create Checklist from Template
**As a** user
**I want to** create a new checklist by copying a template
**So that** I can quickly start a fresh instance (e.g., new shopping trip, new recipe prep)

**Acceptance Criteria:**
- [ ] From Templates page, user can tap "Create checklist" on a template card
- [ ] Modal prompts for new checklist name (auto-suggest: "Template name - [Date]")
- [ ] New checklist is created with all template items pre-filled
- [ ] New checklist is visible in user's main checklist list
- [ ] User is navigated to the new checklist

### Story 4: Manage Templates (Edit/Delete)
**As a** user
**I want to** edit template names/descriptions or delete templates
**So that** my template library stays organized

**Acceptance Criteria:**
- [ ] Templates page shows all user-created templates (list or grid)
- [ ] User can search templates by name
- [ ] User can tap to edit template name/description
- [ ] User can swipe/long-press to delete template
- [ ] Delete shows confirmation
- [ ] Deleting a template doesn't affect checklists created from it

---

## 4. Feature Scope: MVP v1

### In Scope ✅
- Create templates by saving items from a checklist
- Edit template name/description
- Delete templates
- Apply template to existing checklist with smart duplicate detection
- Create new checklist from template
- Templates page (list, search, manage)
- Real-time SSE events for applied templates
- Max 20 items per template
- Mobile-responsive UI (bottom sheets, compact)

### Out of Scope (v2+) ❌
- Pre-built system templates (Carbonara, Shopping lists, etc.)
- Template categories/tagging
- Duplicate templates (copy a template)
- Template sharing between users
- Template versioning/history
- Favorites/starred templates
- Template usage analytics

---

## 5. Data Model

### New Tables

```sql
-- Store templates created by users
CREATE TABLE TEMPLATE (
  ID BIGSERIAL PRIMARY KEY,
  USER_ID BIGINT NOT NULL REFERENCES "user"(ID),
  NAME VARCHAR(255) NOT NULL,
  DESCRIPTION TEXT,
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW(),
  UPDATED_AT TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE(USER_ID, NAME)  -- No duplicate template names per user
);

-- Template items (simple list of ingredients/items)
CREATE TABLE TEMPLATE_ITEM (
  ID BIGSERIAL PRIMARY KEY,
  TEMPLATE_ID BIGINT NOT NULL REFERENCES TEMPLATE(ID) ON DELETE CASCADE,
  NAME VARCHAR(255) NOT NULL,
  POSITION FLOAT NOT NULL,  -- For ordering (same as ChecklistItem)
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW(),
  UPDATED_AT TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Domain Entities (Go)

```go
// internal/core/domain/template.go

type Template struct {
    ID          uint
    UserID      uint
    Name        string
    Description *string
    Items       []TemplateItem
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type TemplateItem struct {
    ID        uint
    TemplateID uint
    Name      string
    Position  float64
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Duplicate Detection Logic

When applying a template, the backend/frontend:
1. Fetches all items in the target checklist
2. Extracts item names
3. Compares template items against checklist items (case-insensitive)
4. Returns: `{ template items, existing items in checklist, which items are new }`
5. Frontend shows picker, user selects
6. Backend creates only selected items

---

## 6. API Endpoints (OpenAPI)

### Templates Management

```
GET    /api/v1/templates                    - List user's templates (paginated, searchable)
POST   /api/v1/templates                    - Create new template
GET    /api/v1/templates/{templateId}       - Get template details with items
PUT    /api/v1/templates/{templateId}       - Update template (name/description)
DELETE /api/v1/templates/{templateId}       - Delete template
```

### Apply Template

```
POST   /api/v1/checklists/{checklistId}/apply-template/{templateId}
       - Apply template to existing checklist
       - Body: { "itemIds": [1, 3, 5] }  // Which items to add (user's selection)
       - Response: Array of created ChecklistItems

GET    /api/v1/checklists/{checklistId}/template-preview/{templateId}
       - Get template preview for checklist (duplicate detection)
       - Response: {
           "template": { ... },
           "existingItemsInChecklist": ["eggs", "pasta"],
           "newItemsToAdd": ["pancetta", "pecorino", "guanciale"]
         }
```

### Template Creation from Selection

```
POST   /api/v1/templates/from-items
       - Create template from selected checklist items
       - Body: {
           "name": "Carbonara",
           "description": "Classic pasta recipe",
           "itemIds": [10, 11, 12]  // Checklist item IDs to copy
         }
       - Response: Created Template

POST   /api/v1/templates/{templateId}/create-checklist
       - Create new checklist from template
       - Body: { "name": "My Shopping - Mar 1" }
       - Response: Created Checklist with pre-populated items
```

---

## 7. Frontend Flows

### Flow 1: Save Items as Template
```
Checklist (viewing items)
  ↓
Long-press item or menu → "Select items to save as template"
  ↓
Checkboxes appear on items
  ↓
User selects multiple items (e.g., eggs, pasta, pancetta)
  ↓
"Save as template" button
  ↓
Modal: Enter template name + optional description
  ↓
Save button → confirmation toast
  ↓
Template saved to library
```

### Flow 2: Apply Template (with Duplicate Detection)
```
Checklist (viewing items)
  ↓
Floating action button: "Add from template" or menu
  ↓
Bottom sheet modal: Template picker (searchable list)
  ↓
User selects template (e.g., "Carbonara")
  ↓
Preview shows:
  - Carbonara items: eggs ✓ (you have), pasta ✓ (you have), pancetta, pecorino
  ↓
User unchecks "eggs" and "pasta"
  ↓
"Add items" button
  ↓
New items (pancetta, pecorino) are added to checklist
  ↓
Real-time update for all connected clients
```

### Flow 3: Create Checklist from Template
```
Templates page
  ↓
Template card: swipe/menu → "Create checklist"
  ↓
Modal: Enter new checklist name
  ↓
"Create" button
  ↓
Navigate to new checklist (all template items pre-filled)
```

### Flow 4: Browse/Manage Templates
```
Templates tab/page
  ↓
List of all user templates (grid or list)
  ↓
Search bar at top (filter by name)
  ↓
Template card actions:
  - Tap: view details + description
  - Swipe/long-press: edit, delete, create checklist
  ↓
Edit modal: change name/description
  ↓
Delete: confirmation dialog
```

---

## 8. Mobile UI Patterns

### Key Components
- **Bottom sheet modal** for template picker (Apple drag handle)
- **Floating action button** for "Add from template" in checklist
- **Checkbox list** for duplicate detection UI
- **Swipe/long-press actions** for delete, edit, create checklist
- **Search bar** at top of Templates page
- **Toast notifications** for confirmations

### Compact Design
- No full-page modals unless necessary
- Scrollable lists (not paginated tabs)
- Touch-friendly button sizing (min 44x44pt)
- **Descriptions shown in collapsed preview** (expandable)

---

## 9. Edge Cases & Constraints

| Case | Decision |
|------|----------|
| **Duplicate template names** | Not allowed per user (DB constraint: UNIQUE) |
| **Template with >20 items** | Warn user at creation time, allow but note limit |
| **Empty template** | Require ≥1 item (prevent save) |
| **Deleting a template** | Doesn't affect checklists created from it (independent copy) |
| **Applying template to shared checklist** | Guard rail: only checklist owner can apply templates |
| **Applying same template twice** | Allowed (user may want duplicates intentionally) |
| **Item name with special chars** | Allow (same as ChecklistItem) |
| **Search case-sensitivity** | Case-insensitive search by name |
| **Concurrent template edits** | Last-write-wins (typical REST behavior) |
| **Template with 0 remaining items** | If user unchecks all items in preview, show warning "no items to add?" |

---

## 10. Success Metrics (Post-Launch)

- % of users who create ≥1 template (adoption)
- Avg template reuse rate (how often applied)
- User satisfaction with duplicate detection (did it help?)
- Time to apply template vs. manual entry (should be <1 minute)

---

## 11. Implementation Phases

### Phase 1: Backend (Core) — 1-2 weeks
1. Create TEMPLATE and TEMPLATE_ITEM tables
2. Domain entities + repository interfaces
3. Template service (CRUD, apply, duplicate detection)
4. OpenAPI spec + code generation
5. Controllers (all endpoints)
6. Unit tests

### Phase 2: Frontend (Core) — 1-2 weeks
1. Templates page (list, search, manage)
2. "Apply template" flow (picker → duplicate detection → confirm)
3. "Save items as template" from checklist
4. "Create checklist from template"
5. Real-time updates (SSE)

### Phase 3: Polish — 3-5 days
1. Mobile UX audit (responsiveness, touch targets)
2. Error handling + edge cases
3. Loading states, empty states
4. Tests (integration, e2e)

---

## 12. Deliverables

### Backend
- ✅ Product plan (this document)
- ⬜ Database schema (SQL)
- ⬜ Domain entities (Go)
- ⬜ Repository interface + implementation
- ⬜ Service layer (business logic)
- ⬜ OpenAPI spec updates
- ⬜ Controllers/handlers
- ⬜ Tests

### Frontend
- ✅ Design documentation (TEMPLATES_DESIGN.md)
- ✅ React components (5 components)
  - TemplatesPage
  - TemplateCard
  - SaveTemplateModal
  - ApplyTemplateSheet (hero component)
  - TemplateIntegration
- ⬜ Integration with API
- ⬜ Real-time SSE integration
- ⬜ Tests

---

## 13. Next Steps

1. ✅ **Product plan complete & confirmed**
2. ✅ **Frontend designs created** (see TEMPLATES_DESIGN.md + React components)
3. ⬜ **Backend implementation** (`/feature` command)
4. ⬜ **Frontend integration** (connect to API)
5. ⬜ **Testing & launch** (MVP v1 complete)

---

## Appendix: Design System Reference

See [TEMPLATES_DESIGN.md](../studio/TEMPLATES_DESIGN.md) for:
- Aesthetic direction and visual guidelines
- Component specifications and props
- User flows and interaction patterns
- Mobile-first optimizations
- Animation and micro-interactions
- Accessibility standards
- API integration points
- Testing checklist

---
