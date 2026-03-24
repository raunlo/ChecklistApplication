# Work Handoff: Templates Feature — Frontend

## Current Status

**Backend**: DONE
**Frontend**: TODO
**Branch**: `templates-uus`
**Last Updated**: 2026-03-25

## Mida template tähendab

**1 template = 1 checklist item + selle row'd.**

Template on blueprint. Kui rakendad template'i checklistile, tekib 1 uus ChecklistItem koos oma row'dega. Template EI ole checklist ise ega kogum itemeid.

**Näide:** Template "Hommikusöök" (rows: "Kohv", "Munad", "Leib") → rakendamisel tekib checklist'i üks item "Hommikusöök" millel on 3 row'd.

---

## Backend API endpointid

| Meetod | Endpoint | Kirjeldus |
|--------|----------|-----------|
| GET | `/api/v1/templates` | Kõik kasutaja template'd |
| POST | `/api/v1/templates` | Loo uus template (nimi + optional rows) |
| GET | `/api/v1/templates/{id}` | Üks template koos row'dega |
| PUT | `/api/v1/templates/{id}` | Uuenda template nime/kirjeldust |
| DELETE | `/api/v1/templates/{id}` | Kustuta template |
| POST | `/api/v1/templates/from-items` | Loo template olemasolevast checklist itemist |
| POST | `/api/v1/checklists/{id}/apply-template/{templateId}` | Rakenda template → loob 1 item + rows |

### Request/Response näited

**Loo template:**
```json
POST /api/v1/templates
{
  "name": "Hommikusöök",
  "description": "Hommikune rutiin",
  "rows": [
    { "name": "Kohv", "position": 1000 },
    { "name": "Munad", "position": 2000 },
    { "name": "Leib", "position": 3000 }
  ]
}
```

**Rakenda template checklistile (no body!):**
```json
POST /api/v1/checklists/33/apply-template/5
→ Returns: ChecklistItemResponse (üks item koos row'dega)
```

**Loo template olemasolevast itemist:**
```json
POST /api/v1/templates/from-items
{
  "name": "Minu template",
  "checklistId": 33,
  "checklistItemId": 42
}
```

---

## Frontend UX plaan

### Peamine idee: Template drawer checklist detail lehel

Checklist detail lehel (`/checklist/[id]`) peaks olema **alt üles tõmmatav drawer/sheet** kus kasutaja saab template'eid sirvida ja rakendada.

#### Flow:

```
Checklist detail leht
        ↓
Kasutaja tõmbab alt üles (või vajutab nuppu)
        ↓
┌─────────────────────────────────┐
│  📋 Templates                   │
│                                 │
│  ┌───────────────────────────┐  │
│  │ Hommikusöök          (3)  │  │  ← tap = rakenda
│  │ Kohv, Munad, Leib        │  │
│  └───────────────────────────┘  │
│  ┌───────────────────────────┐  │
│  │ Puhastus             (5)  │  │
│  │ Tolm, Põrand, Aken...    │  │
│  └───────────────────────────┘  │
│                                 │
│  [+ Loo uus template]          │
│                                 │
│  ──── swipe down to close ────  │
└─────────────────────────────────┘
        ↓
Kasutaja tap'ib template'i
        ↓
Confirm dialog: "Lisa 'Hommikusöök' selle checklisti?"
        ↓
POST /api/v1/checklists/{id}/apply-template/{templateId}
        ↓
Uus item ilmub checklisti (SSE event)
```

### Teine entry point: "Save as Template" olemasolevalt itemilt

Checklist detail lehel, iga itemi kolm-punkti menüüs (või long-press):
- "Save as Template" → `POST /api/v1/templates/from-items`
- Küsib nime → loob template itemi row'dest

### Templates management leht (`/templates`)

Jääb alles kui secondary management leht:
- Nimeta ümber, kustuta template'eid
- Lisa/eemalda row'sid
- **EI ole** peamine koht kust template'eid kasutada

---

## Konkreetsed ülesanded frontendile

### 1. API hookid (uuenda `src/api/template/template.ts`)

Praegused hookid on VALED (vana mudel). Uuenda:
- `useGetAllTemplates()` — GET /api/v1/templates
- `useGetTemplateById(id)` — GET /api/v1/templates/{id}
- `useCreateTemplate()` — POST /api/v1/templates
- `useUpdateTemplate()` — PUT /api/v1/templates/{id}
- `useDeleteTemplate()` — DELETE /api/v1/templates/{id}
- `useApplyTemplate()` — POST /api/v1/checklists/{checklistId}/apply-template/{templateId} **(no body!)**
- `useCreateTemplateFromItem()` — POST /api/v1/templates/from-items

### 2. Template drawer komponent (UUS)

`src/components/template-drawer.tsx`:
- Bottom sheet/drawer UI (Radix `Sheet` component, suund: bottom)
- Näitab kõik template'd (useGetAllTemplates)
- Iga template card: nimi, row'de arv, row'de eelvaade
- Tap → confirm dialog → applyTemplate
- "+ Loo uus" nupp → lühike form (nimi + kirjeldus)

### 3. Checklist detail lehel integratsioon

`src/app/checklist/[id]/page.tsx`:
- Asenda praegune "Templates" nupp (Zap icon) → avab template drawer'i
- Alternatiiv: lisa pull-up gesture handle ekraani alumisse serva
- Pärast apply't → SWR mutate et uus item ilmuks

### 4. "Save as Template" toiming

`src/components/checklist-item.tsx` (või kus iganes itemi context menu on):
- Lisa menüü-option "Save as Template"
- Avab väikse dialoogi: template nimi
- Kutsub `useCreateTemplateFromItem({ name, checklistId, checklistItemId })`

### 5. Templates management leht (optionaalne koristus)

`src/app/templates/page.tsx` + `src/components/template-overview.tsx`:
- Eemalda TemplateQuickGuide (enam pole vaja)
- Uuenda mudel: `items` → `rows`
- Lisa "Back to Checklists" link

---

## TypeScript tüübid (uus mudel)

```typescript
interface Template {
  id: number;
  userId: string;
  name: string;
  description?: string;
  rows: TemplateRow[];
  createdAt: string;
  updatedAt: string;
}

interface TemplateRow {
  id: number;
  templateId: number;
  name: string;
  position: number;
  createdAt: string;
  updatedAt: string;
}

interface CreateTemplateRequest {
  name: string;
  description?: string;
  rows?: { name: string; position: number }[];
}

interface CreateTemplateFromItemRequest {
  name: string;
  description?: string;
  checklistId: number;
  checklistItemId: number;
}
```

---

## Mida MITTE teha

- Ära tee template'ist top-level nav itemit
- Ära lisa tutoriale/guide'e — UX peab olema selge ilma
- Ära loo template'ist uut checklisti (seda endpointi enam pole)
- Ära kasuta vana `ApplyTemplateRequest` body't (enam pole itemIds valikut)

---

## Prioriteedid

1. **Template drawer** checklist detail lehel (peamine UX)
2. **API hookide** uuendus (vanad on katki)
3. **"Save as Template"** itemi menüüst
4. Templates management lehe koristus (madal prioriteet)
