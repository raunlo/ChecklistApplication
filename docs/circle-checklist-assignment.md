# Circle & Checklist Assignment Feature

## What Was Built

Two related features were added:

1. **Assign existing checklists to circles** — from the circle detail view, users can pick an owned checklist and move it into that circle.
2. **Unified "Edit checklist" modal** — replaces separate "Rename" and "Add to circle" dropdown items with a single Edit dialog that manages both name and circle assignment in one step.

---

## Backend Fix

**File:** `internal/repository/checklist_repository.go` — `UpdateChecklist`

The SQL query only updated `NAME`, not `workspace_id`:

```sql
-- Before (broken):
UPDATE checklist SET NAME = @checklist_name WHERE ID = @checklist_id

-- After (fixed):
UPDATE checklist
SET NAME = @checklist_name, workspace_id = @workspace_id
WHERE ID = @checklist_id
```

The Go args were also updated to include `"workspace_id": checklist.WorkspaceId`.

The frontend was already sending `workspaceId` in the PUT body via `updateChecklistById(id, { name, workspaceId })` — the backend just wasn't persisting it.

---

## Frontend Changes

### 1. Shared color utility

`src/lib/circle-colors.ts` already existed. Both `workspace-overview.tsx` and `checklist-overview.tsx` import `getCircleColor(index)` from here. Do not duplicate the `CIRCLE_COLORS` array.

### 2. Edit Checklist Modal (`checklist-overview.tsx`)

Replaced two separate dialogs (Rename + Assign to Circle) with one unified modal.

**State:**
```tsx
const [editDialogOpen, setEditDialogOpen] = useState(false);
const [editingChecklist, setEditingChecklist] = useState<{
  id: number; name: string; workspaceId?: number | null
} | null>(null);
const [editName, setEditName] = useState('');
const [editWorkspaceId, setEditWorkspaceId] = useState<number | null>(null);
const [isSavingEdit, setIsSavingEdit] = useState(false);
```

**Open handler** — finds the checklist from SWR data, pre-fills name and current circle:
```tsx
const handleOpenEdit = (id: number) => {
  const found = checklists.find((c) => c.id === id);
  if (!found) return;
  setEditingChecklist({ id: found.id, name: found.name, workspaceId: found.workspaceId });
  setEditName(found.name);
  setEditWorkspaceId(found.workspaceId ?? null);
  setEditDialogOpen(true);
};
```

**Save handler** — calls PUT endpoint, then invalidates SWR cache:
```tsx
const handleSaveEdit = async () => {
  if (!editingChecklist || !editName.trim()) return;
  setIsSavingEdit(true);
  try {
    await updateChecklistById(editingChecklist.id, {
      name: editName.trim(),
      workspaceId: editWorkspaceId ?? undefined,
    });
    mutate(getGetAllChecklistsKey());
    setEditDialogOpen(false);
  } catch {
    toast({ title: t('toast.failedToSave'), variant: 'destructive' });
  } finally {
    setIsSavingEdit(false);
  }
};
```

**Circle chips UI** — colored dots matching the circles view, tap to toggle:
```tsx
{ownedWorkspaces.map((w, i) => {
  const color = getCircleColor(i);
  const selected = editWorkspaceId === w.id;
  return (
    <button
      key={w.id}
      type="button"
      onClick={() => setEditWorkspaceId(selected ? null : w.id)}
      className={`flex items-center gap-1.5 rounded-full border px-3 py-1 text-sm transition-colors ${
        selected
          ? 'text-foreground'
          : 'border-border text-muted-foreground hover:text-foreground'
      }`}
      style={selected ? { backgroundColor: color + '22', borderColor: color + '88' } : {}}
    >
      <span className="h-2 w-2 flex-shrink-0 rounded-full" style={{ backgroundColor: color }} />
      {w.name}
    </button>
  );
})}
```

Tapping a selected chip deselects it (removes from circle). The dialog also uses `flex justify-end gap-3` instead of `DialogFooter` to keep buttons side-by-side on mobile.

### 3. Card dropdown (`checklist-overview-card.tsx`)

The `onAssignToCircle` prop and `FolderPlus` icon were removed. "Rename" was renamed to "Edit" — it still calls `onEdit`. The card now only shows Edit + Delete (or Leave for non-owners).

### 4. Assign existing checklist to circle (`workspace-overview-detail.tsx`)

An "Add existing" button was added next to "New checklist" in the circle detail's checklists tab. It opens a dialog with a search input and a scrollable list of the user's owned checklists not already in this circle.

**Filter logic:**
```tsx
const inCircleIds = useMemo(() => new Set(circleChecklists.map((c) => c.id)), [circleChecklists]);
const assignableChecklists = useMemo(
  () => allChecklists.filter((c) => c.isOwner && !inCircleIds.has(c.id)),
  [allChecklists, inCircleIds]
);
```

This allows moving a checklist from one circle to another (not just from "no circle").

**Assign handler:**
```tsx
const handleAssignChecklist = async (checklist: ChecklistResponse) => {
  await updateChecklistById(checklist.id, {
    name: checklist.name,
    workspaceId: workspaceId,
  });
  mutate(getGetWorkspaceChecklistsKey(workspaceId));
  setAssignDialogOpen(false);
};
```

---

## Key Points

- `workspaceId: null` removes a checklist from its circle (moves to personal/default workspace).
- `workspaceId: <id>` assigns it to that circle.
- The backend `UpdateChecklist` now always persists both name and workspace_id together.
- Colors are driven by workspace index in the list, not a stored property — same index = same color across views.
- Use inline `style` for dynamic hex colors; Tailwind can't handle runtime values.
- Dialog buttons: use `<div className="flex justify-end gap-3 pt-2">` not `<DialogFooter>` to prevent mobile stacking.
