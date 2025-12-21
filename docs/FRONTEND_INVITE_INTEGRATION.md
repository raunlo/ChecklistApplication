# Frontend Integration Guide: Shareable Checklist Invites

**Version:** 1.0
**Last Updated:** 2025-01-21
**Backend API Version:** v1

## Overview

This document describes how to implement the shareable invite link feature in the frontend. The feature allows checklist owners to generate invite links that can be shared via any channel (WhatsApp, email, Slack, etc.). Recipients claim the invite to gain access to the checklist.

**Privacy Note:** The system stores zero PII (no emails, names). Only Google IDs from JWT tokens are stored.

---

## Table of Contents

1. [API Endpoints](#api-endpoints)
2. [Data Models](#data-models)
3. [User Flows](#user-flows)
4. [Component Specifications](#component-specifications)
5. [Code Examples](#code-examples)
6. [Error Handling](#error-handling)
7. [Edge Cases](#edge-cases)
8. [Testing Checklist](#testing-checklist)

---

## API Endpoints

### 1. Create Invite Link

**Endpoint:** `POST /api/v1/checklists/{checklistId}/invites`

**Authorization:** Owner only (returns 403 if user is not checklist owner)

**Request:**
```typescript
interface CreateInviteRequest {
  expiresInHours?: number | null;  // Hours until expiration (null = never expires)
  isSingleUse: boolean;             // If true, can only be claimed once
}
```

**Request Example:**
```javascript
POST /api/v1/checklists/123/invites
Headers:
  X-Client-Id: abc-123-client-uuid
  Cookie: user_token=...
Body:
{
  "expiresInHours": 168,  // 7 days
  "isSingleUse": true
}
```

**Response (201 Created):**
```typescript
interface InviteResponse {
  id: number;
  checklistId: number;
  inviteToken: string;          // 64-character secure token
  inviteUrl: string;            // Full URL for sharing: "https://app.com/invites/{token}/claim"
  createdAt: string;            // ISO 8601 timestamp
  expiresAt: string | null;     // ISO 8601 timestamp or null if never expires
  claimedBy: string | null;     // Google ID if claimed (don't show to user, use for logic only)
  claimedAt: string | null;     // ISO 8601 timestamp or null
  isSingleUse: boolean;
  isExpired: boolean;           // Computed: true if past expiresAt
  isClaimed: boolean;           // Computed: true if claimedAt is set
}
```

**Response Example:**
```json
{
  "id": 456,
  "checklistId": 123,
  "inviteToken": "a1b2c3d4e5f6...",
  "inviteUrl": "https://app.dailychexly.com/invites/a1b2c3d4e5f6.../claim",
  "createdAt": "2025-01-21T10:00:00Z",
  "expiresAt": "2025-01-28T10:00:00Z",
  "claimedBy": null,
  "claimedAt": null,
  "isSingleUse": true,
  "isExpired": false,
  "isClaimed": false
}
```

**Error Responses:**
- `403 Forbidden` - User is not the checklist owner
- `404 Not Found` - Checklist doesn't exist or user lacks access
- `400 Bad Request` - Invalid request body (e.g., expiresInHours < 1)
- `500 Internal Server Error` - Server error

---

### 2. List Active Invites

**Endpoint:** `GET /api/v1/checklists/{checklistId}/invites`

**Authorization:** Owner only

**Response (200 OK):**
```typescript
type InviteListResponse = InviteResponse[];
```

**Response Example:**
```json
[
  {
    "id": 456,
    "inviteToken": "a1b2c3...",
    "inviteUrl": "https://app.dailychexly.com/invites/a1b2c3.../claim",
    "createdAt": "2025-01-21T10:00:00Z",
    "expiresAt": "2025-01-28T10:00:00Z",
    "claimedAt": null,
    "isSingleUse": true,
    "isExpired": false,
    "isClaimed": false
  },
  {
    "id": 457,
    "inviteToken": "x9y8z7...",
    "inviteUrl": "https://app.dailychexly.com/invites/x9y8z7.../claim",
    "createdAt": "2025-01-20T14:00:00Z",
    "expiresAt": null,
    "claimedAt": "2025-01-20T15:30:00Z",
    "isSingleUse": false,
    "isExpired": false,
    "isClaimed": true
  }
]
```

**Error Responses:**
- `403 Forbidden` - User is not the checklist owner
- `404 Not Found` - Checklist doesn't exist

---

### 3. Revoke Invite

**Endpoint:** `DELETE /api/v1/checklists/invites/{inviteId}`

**Authorization:** Owner of the checklist associated with the invite

**Response (204 No Content):** Empty body

**Error Responses:**
- `403 Forbidden` - User is not the owner of the checklist this invite belongs to
- `404 Not Found` - Invite doesn't exist
- `500 Internal Server Error` - Server error

---

### 4. Claim Invite

**Endpoint:** `POST /api/v1/invites/{token}/claim`

**Authorization:** Any authenticated user

**Request:** Empty body

**Request Example:**
```javascript
POST /api/v1/invites/a1b2c3d4e5f6.../claim
Headers:
  X-Client-Id: abc-123-client-uuid
  Cookie: user_token=...
```

**Response (200 OK):**
```typescript
interface ClaimInviteResponse {
  checklistId: number;
  message: string;
}
```

**Response Example:**
```json
{
  "checklistId": 123,
  "message": "Successfully joined checklist"
}
```

**Error Responses:**
- `400 Bad Request` - Invite expired, already claimed (single-use), or invalid
  ```json
  { "error": "Invite expired", "message": "This invite link has expired" }
  ```
  ```json
  { "error": "Invite already claimed", "message": "This invite has already been used" }
  ```
- `404 Not Found` - Invite doesn't exist
  ```json
  { "error": "Invite not found", "message": "This invite link is invalid" }
  ```
- `401 Unauthorized` - User not authenticated
- `500 Internal Server Error` - Server error

**Important:** This endpoint is **idempotent**. If the user already has access to the checklist, it returns success (200) with the checklist ID. This prevents errors if a user clicks the same link twice.

---

## Data Models

### TypeScript Interfaces

```typescript
// Request/Response types
interface CreateInviteRequest {
  expiresInHours?: number | null;
  isSingleUse: boolean;
}

interface InviteResponse {
  id: number;
  checklistId: number;
  inviteToken: string;
  inviteUrl: string;
  createdAt: string;
  expiresAt: string | null;
  claimedBy: string | null;  // Don't display to users (privacy)
  claimedAt: string | null;
  isSingleUse: boolean;
  isExpired: boolean;
  isClaimed: boolean;
}

interface ClaimInviteResponse {
  checklistId: number;
  message: string;
}

// UI State types
type InviteStatus = 'active' | 'claimed' | 'expired';

interface InviteDisplayData {
  id: number;
  url: string;
  status: InviteStatus;
  createdAt: Date;
  expiresAt: Date | null;
  claimedAt: Date | null;
  isSingleUse: boolean;
  expiryLabel: string;  // "Expires in 2 days" | "Never expires" | "Expired"
  statusLabel: string;  // "Active" | "Claimed" | "Expired"
}
```

### Helper Functions

```typescript
// Convert API response to display-friendly format
function mapInviteToDisplay(invite: InviteResponse): InviteDisplayData {
  const now = new Date();
  const expiresAt = invite.expiresAt ? new Date(invite.expiresAt) : null;
  const claimedAt = invite.claimedAt ? new Date(invite.claimedAt) : null;

  let status: InviteStatus;
  if (invite.isExpired) {
    status = 'expired';
  } else if (invite.isClaimed && invite.isSingleUse) {
    status = 'claimed';
  } else {
    status = 'active';
  }

  return {
    id: invite.id,
    url: invite.inviteUrl,
    status,
    createdAt: new Date(invite.createdAt),
    expiresAt,
    claimedAt,
    isSingleUse: invite.isSingleUse,
    expiryLabel: getExpiryLabel(expiresAt, invite.isExpired),
    statusLabel: getStatusLabel(status),
  };
}

function getExpiryLabel(expiresAt: Date | null, isExpired: boolean): string {
  if (!expiresAt) return 'Never expires';
  if (isExpired) return 'Expired';

  const now = new Date();
  const diff = expiresAt.getTime() - now.getTime();
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24));
  const hours = Math.ceil(diff / (1000 * 60 * 60));

  if (days > 1) return `Expires in ${days} days`;
  if (hours > 1) return `Expires in ${hours} hours`;
  return 'Expires soon';
}

function getStatusLabel(status: InviteStatus): string {
  switch (status) {
    case 'active': return 'Active';
    case 'claimed': return 'Claimed';
    case 'expired': return 'Expired';
  }
}
```

---

## User Flows

### Flow 1: Owner Creates and Shares Invite

```
1. User navigates to "My Checklists" page
2. User clicks three-dot menu (⋮) on a checklist card
3. Dropdown shows ["Edit", "Share", "Delete"]
4. User clicks "Share"
   → ShareChecklistModal opens
   → API call: GET /api/v1/checklists/{id}/invites
   → Modal displays list of active invites

5. User configures new invite:
   - Checks "Single use only" checkbox (default: true)
   - Selects expiration from dropdown:
     • Never (expiresInHours: null)
     • 1 day (expiresInHours: 24)
     • 7 days (expiresInHours: 168)
     • 30 days (expiresInHours: 720)

6. User clicks "Generate Link"
   → API call: POST /api/v1/checklists/{id}/invites
   → Response contains inviteUrl
   → Automatically copy inviteUrl to clipboard
   → Show toast: "Invite link copied to clipboard!"
   → Refresh invite list

7. User shares the copied URL via WhatsApp/Email/Slack/etc.

8. (Optional) User clicks "Copy" on existing invite
   → Copy inviteUrl to clipboard
   → Show toast: "Link copied!"

9. (Optional) User clicks "Revoke" on an invite
   → Show confirmation: "Revoke this invite link?"
   → API call: DELETE /api/v1/checklists/invites/{inviteId}
   → Refresh invite list
   → Show toast: "Invite revoked"
```

### Flow 2: Recipient Claims Invite

```
1. Recipient receives URL: https://app.dailychexly.com/invites/{token}/claim
   (via WhatsApp, email, Slack, etc.)

2. Recipient clicks the link

3. Frontend checks authentication:
   IF not authenticated:
     → Redirect to /login?returnUrl=/invites/{token}/claim
     → User logs in with Google OAuth
     → After successful login, redirect to /invites/{token}/claim

   IF authenticated:
     → Continue to step 4

4. ClaimInvitePage component loads
   → Show loading spinner: "Joining checklist..."
   → Automatically call API: POST /api/v1/invites/{token}/claim

5. Handle API response:

   SUCCESS (200):
     → Hide spinner
     → Show success message: "✓ You've joined the checklist!"
     → Auto-redirect to /checklists/{checklistId} after 1.5 seconds
     → User now has access to the checklist

   ERROR (400 - Expired):
     → Show error: "This invite link has expired"
     → Show button: "Back to Home"

   ERROR (400 - Already Claimed):
     → Show error: "This invite has already been used"
     → Show button: "Back to Home"

   ERROR (404):
     → Show error: "Invalid invite link"
     → Show button: "Back to Home"

6. After successful claim:
   → Recipient goes to "My Checklists"
   → Shared checklist appears in "Shared with me" section
   → Checklist has 🔗 icon to indicate it's shared
   → Recipient can open checklist normally
   → Real-time SSE updates work automatically
```

### Flow 3: Viewing Shared Checklist

```
1. After claiming invite, recipient sees checklist in "My Checklists"
   → GET /api/v1/checklists returns both owned and shared checklists
   → Backend query: WHERE c.OWNER = @user_id OR cs.SHARED_WITH_USER_ID = @user_id

2. Frontend groups checklists:
   Owned (isOwner: true):
     - Can edit, delete, share
     - Three-dot menu shows all options

   Shared (isOwner: false):
     - Can view and edit items
     - Three-dot menu hides "Share" and "Delete"
     - Shows 🔗 icon to indicate shared status

3. User clicks shared checklist
   → Opens checklist detail page normally
   → GET /api/v1/checklists/{id} works (guard rail checks CHECKLIST_SHARE)
   → GET /api/v1/checklists/{id}/items works
   → SSE connection works (guard rail on subscribe checks CHECKLIST_SHARE)

4. Real-time collaboration works automatically:
   → Owner and recipient both have checklist open
   → Owner adds item → SSE event → Recipient's UI updates
   → Recipient checks item → SSE event → Owner's UI updates
   → X-Client-Id prevents echo (user doesn't see their own changes twice)
```

---

## Component Specifications

### Component 1: ShareChecklistModal

**Purpose:** Allow checklist owner to manage invite links

**Props:**
```typescript
interface ShareChecklistModalProps {
  checklistId: number;
  checklistName: string;
  isOpen: boolean;
  onClose: () => void;
}
```

**State:**
```typescript
interface ShareChecklistModalState {
  invites: InviteDisplayData[];
  isLoading: boolean;
  isCreating: boolean;
  error: string | null;

  // Form state
  selectedExpiry: 'never' | '1day' | '7days' | '30days';
  isSingleUse: boolean;
}
```

**UI Layout:**
```
┌─────────────────────────────────────────────────┐
│ Share "Grocery List"                      [X]   │
├─────────────────────────────────────────────────┤
│                                                 │
│ Create New Invite Link                         │
│ ┌─────────────────────────────────────────────┐│
│ │ ☑ Single use only                          ││
│ │                                             ││
│ │ Expires in: [Dropdown ▼]                   ││
│ │   • Never                                   ││
│ │   • 1 day                                   ││
│ │   • 7 days                                  ││
│ │   • 30 days                                 ││
│ │                                             ││
│ │ [Generate Invite Link]                     ││
│ └─────────────────────────────────────────────┘│
│                                                 │
│ Active Invites (3)                             │
│ ┌─────────────────────────────────────────────┐│
│ │ 🔗 https://app.com/invites/abc...           ││
│ │    Expires in 5 days • Single use           ││
│ │    [Copy Link] [Revoke]                     ││
│ ├─────────────────────────────────────────────┤│
│ │ 🔗 https://app.com/invites/xyz...           ││
│ │    Never expires • Reusable                 ││
│ │    Claimed on Jan 20 at 3:30 PM             ││
│ │    [Copy Link] [Revoke]                     ││
│ ├─────────────────────────────────────────────┤│
│ │ 🔗 https://app.com/invites/def...           ││
│ │    Expired                                  ││
│ │    [Revoke]                                 ││
│ └─────────────────────────────────────────────┘│
│                                                 │
│                                   [Close]       │
└─────────────────────────────────────────────────┘
```

**Behavior:**

1. **On Open:**
   - Fetch active invites: `GET /api/v1/checklists/{id}/invites`
   - Display loading spinner while fetching
   - Default form: `isSingleUse = true`, `selectedExpiry = '7days'`

2. **Generate Link Button:**
   - Map selectedExpiry to hours: `{ never: null, '1day': 24, '7days': 168, '30days': 720 }`
   - Call API: `POST /api/v1/checklists/{id}/invites`
   - On success:
     - Copy `inviteUrl` to clipboard using `navigator.clipboard.writeText()`
     - Show toast notification: "Invite link copied to clipboard!"
     - Refresh invite list
     - Reset form to defaults
   - On error:
     - Show error message in modal
     - If 403: "You don't have permission to create invites"

3. **Copy Link Button:**
   - Copy invite's `inviteUrl` to clipboard
   - Show toast: "Link copied!"

4. **Revoke Button:**
   - Show confirmation dialog: "Are you sure you want to revoke this invite link? Anyone with this link will no longer be able to use it."
   - If confirmed:
     - Call API: `DELETE /api/v1/checklists/invites/{inviteId}`
     - On success: Refresh invite list, show toast "Invite revoked"
     - On error: Show error message

5. **Display Logic:**
   - Filter out expired invites from the list (optional, or show grayed out)
   - Sort by: active first, then by creation date (newest first)
   - For claimed single-use invites, show "Claimed on [date]" instead of action buttons
   - Truncate long URLs in display, show full URL on hover or in tooltip

**Example Code:**
```typescript
const ShareChecklistModal: React.FC<ShareChecklistModalProps> = ({
  checklistId,
  checklistName,
  isOpen,
  onClose,
}) => {
  const [invites, setInvites] = useState<InviteDisplayData[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedExpiry, setSelectedExpiry] = useState<'never' | '1day' | '7days' | '30days'>('7days');
  const [isSingleUse, setIsSingleUse] = useState(true);

  useEffect(() => {
    if (isOpen) {
      fetchInvites();
    }
  }, [isOpen]);

  const fetchInvites = async () => {
    setIsLoading(true);
    try {
      const response = await fetch(`/api/v1/checklists/${checklistId}/invites`, {
        headers: { 'X-Client-Id': getClientId() },
        credentials: 'include',
      });

      if (response.ok) {
        const data: InviteResponse[] = await response.json();
        setInvites(data.map(mapInviteToDisplay));
      }
    } catch (err) {
      console.error('Failed to fetch invites:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleGenerateLink = async () => {
    const expiryMap = { never: null, '1day': 24, '7days': 168, '30days': 720 };
    const expiresInHours = expiryMap[selectedExpiry];

    try {
      const response = await fetch(`/api/v1/checklists/${checklistId}/invites`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Client-Id': getClientId(),
        },
        credentials: 'include',
        body: JSON.stringify({ expiresInHours, isSingleUse }),
      });

      if (response.ok) {
        const invite: InviteResponse = await response.json();

        // Copy to clipboard
        await navigator.clipboard.writeText(invite.inviteUrl);
        showToast('Invite link copied to clipboard!');

        // Refresh list
        fetchInvites();

        // Reset form
        setIsSingleUse(true);
        setSelectedExpiry('7days');
      } else {
        const error = await response.json();
        showToast(error.message || 'Failed to generate invite', 'error');
      }
    } catch (err) {
      showToast('Failed to generate invite', 'error');
    }
  };

  const handleCopyLink = async (url: string) => {
    await navigator.clipboard.writeText(url);
    showToast('Link copied!');
  };

  const handleRevoke = async (inviteId: number) => {
    if (!confirm('Are you sure you want to revoke this invite link?')) return;

    try {
      const response = await fetch(`/api/v1/checklists/invites/${inviteId}`, {
        method: 'DELETE',
        headers: { 'X-Client-Id': getClientId() },
        credentials: 'include',
      });

      if (response.ok) {
        showToast('Invite revoked');
        fetchInvites();
      }
    } catch (err) {
      showToast('Failed to revoke invite', 'error');
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      {/* UI implementation */}
    </Modal>
  );
};
```

---

### Component 2: ClaimInvitePage

**Purpose:** Automatically claim an invite when user visits the link

**Route:** `/invites/:token/claim`

**Props:** None (uses URL parameter)

**State:**
```typescript
type ClaimStatus = 'loading' | 'success' | 'error';

interface ClaimInvitePageState {
  status: ClaimStatus;
  checklistId: number | null;
  errorMessage: string | null;
}
```

**UI States:**

**Loading State:**
```
┌─────────────────────────────────────┐
│                                     │
│         [Loading Spinner]           │
│                                     │
│     Joining checklist...            │
│                                     │
└─────────────────────────────────────┘
```

**Success State:**
```
┌─────────────────────────────────────┐
│                                     │
│              ✓                      │
│                                     │
│   You've joined the checklist!      │
│                                     │
│   Redirecting...                    │
│                                     │
└─────────────────────────────────────┘
```

**Error State (Expired):**
```
┌─────────────────────────────────────┐
│                                     │
│              ✗                      │
│                                     │
│   Invite Link Expired               │
│                                     │
│   This invite link has expired.     │
│   Please request a new invite.      │
│                                     │
│   [Back to Home]                    │
│                                     │
└─────────────────────────────────────┘
```

**Error State (Already Claimed):**
```
┌─────────────────────────────────────┐
│                                     │
│              ✗                      │
│                                     │
│   Invite Already Used               │
│                                     │
│   This invite can only be used      │
│   once and has already been         │
│   claimed.                          │
│                                     │
│   [Back to Home]                    │
│                                     │
└─────────────────────────────────────┘
```

**Error State (Invalid):**
```
┌─────────────────────────────────────┐
│                                     │
│              ✗                      │
│                                     │
│   Invalid Invite Link               │
│                                     │
│   This invite link is invalid or    │
│   doesn't exist.                    │
│                                     │
│   [Back to Home]                    │
│                                     │
└─────────────────────────────────────┘
```

**Behavior:**

1. **On Component Mount:**
   - Extract `token` from URL parameters
   - Set status to 'loading'
   - Immediately call claim API: `POST /api/v1/invites/{token}/claim`

2. **On Success (200):**
   - Set status to 'success'
   - Store checklistId from response
   - Wait 1.5 seconds (show success message)
   - Auto-redirect: `navigate(`/checklists/${checklistId}`)`

3. **On Error (400/404):**
   - Set status to 'error'
   - Parse error message from response
   - Display appropriate error UI based on error type
   - "Back to Home" button navigates to `/checklists`

4. **Authentication Check:**
   - If API returns 401 Unauthorized
   - Redirect to: `/login?returnUrl=/invites/${token}/claim`
   - After login, user is redirected back to claim page

**Example Code:**
```typescript
const ClaimInvitePage: React.FC = () => {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const [status, setStatus] = useState<ClaimStatus>('loading');
  const [checklistId, setChecklistId] = useState<number | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    if (token) {
      claimInvite(token);
    }
  }, [token]);

  const claimInvite = async (token: string) => {
    try {
      const response = await fetch(`/api/v1/invites/${token}/claim`, {
        method: 'POST',
        headers: { 'X-Client-Id': getClientId() },
        credentials: 'include',
      });

      if (response.ok) {
        const data: ClaimInviteResponse = await response.json();
        setStatus('success');
        setChecklistId(data.checklistId);

        // Auto-redirect after showing success message
        setTimeout(() => {
          navigate(`/checklists/${data.checklistId}`);
        }, 1500);

      } else if (response.status === 401) {
        // Not authenticated - redirect to login
        const returnUrl = encodeURIComponent(`/invites/${token}/claim`);
        navigate(`/login?returnUrl=${returnUrl}`);

      } else {
        // Error (400, 404, 500)
        const error = await response.json();
        setStatus('error');
        setErrorMessage(error.message || 'Failed to claim invite');
      }

    } catch (err) {
      setStatus('error');
      setErrorMessage('Network error. Please try again.');
    }
  };

  if (status === 'loading') {
    return (
      <div className="claim-page">
        <Spinner />
        <p>Joining checklist...</p>
      </div>
    );
  }

  if (status === 'success') {
    return (
      <div className="claim-page">
        <SuccessIcon />
        <h2>You've joined the checklist!</h2>
        <p>Redirecting...</p>
      </div>
    );
  }

  if (status === 'error') {
    return (
      <div className="claim-page">
        <ErrorIcon />
        <h2>Unable to Join Checklist</h2>
        <p>{errorMessage}</p>
        <button onClick={() => navigate('/checklists')}>
          Back to Home
        </button>
      </div>
    );
  }

  return null;
};
```

---

### Component 3: ChecklistCard (Modifications)

**Purpose:** Display checklist with visual indicators for shared status

**Existing Props:**
```typescript
interface ChecklistCardProps {
  checklist: {
    id: number;
    name: string;
    isOwner: boolean;  // NEW field from API
    stats: {
      totalItems: number;
      completedItems: number;
    };
  };
  onEdit: (id: number) => void;
  onDelete: (id: number) => void;
  onShare: (id: number) => void;  // NEW callback
}
```

**UI Layout (Owned Checklist):**
```
┌─────────────────────────────────┐
│ Grocery List               [⋮]  │
│ 5/10 items completed            │
└─────────────────────────────────┘

Three-dot menu:
• Edit
• Share     ← Shows ShareChecklistModal
• Delete
```

**UI Layout (Shared Checklist):**
```
┌─────────────────────────────────┐
│ 🔗 Shopping List           [⋮]  │  ← Icon indicates shared
│ Shared with you                 │  ← Subtitle
│ 3/8 items completed             │
└─────────────────────────────────┘

Three-dot menu:
• Edit       ← Only if writable permission (future)
(No Share or Delete options)
```

**Changes Required:**

1. **Add shared indicator icon:**
   ```tsx
   {!checklist.isOwner && <span className="shared-icon">🔗</span>}
   ```

2. **Conditional menu items:**
   ```tsx
   const menuItems = [
     { label: 'Edit', onClick: () => onEdit(checklist.id) },
     // Only show Share and Delete for owners
     ...(checklist.isOwner ? [
       { label: 'Share', onClick: () => onShare(checklist.id) },
       { label: 'Delete', onClick: () => onDelete(checklist.id) },
     ] : [])
   ];
   ```

3. **Add subtitle for shared checklists:**
   ```tsx
   {!checklist.isOwner && (
     <p className="checklist-subtitle">Shared with you</p>
   )}
   ```

---

### Component 4: ChecklistsPage (Modifications)

**Purpose:** Group checklists into "My Checklists" and "Shared with me"

**UI Layout:**
```
┌─────────────────────────────────────────┐
│ My Checklists                           │
├─────────────────────────────────────────┤
│                                         │
│ ┌─────────────────────┐                │
│ │ Grocery List   [⋮]  │                │
│ │ 5/10 completed      │                │
│ └─────────────────────┘                │
│                                         │
│ ┌─────────────────────┐                │
│ │ Work Tasks     [⋮]  │                │
│ │ 2/5 completed       │                │
│ └─────────────────────┘                │
│                                         │
├─────────────────────────────────────────┤
│ Shared with me                          │
├─────────────────────────────────────────┤
│                                         │
│ ┌─────────────────────┐                │
│ │ 🔗 Shopping List[⋮] │                │
│ │ Shared with you     │                │
│ │ 3/8 completed       │                │
│ └─────────────────────┘                │
│                                         │
└─────────────────────────────────────────┘
```

**Changes Required:**

```typescript
const ChecklistsPage: React.FC = () => {
  const [checklists, setChecklists] = useState<Checklist[]>([]);

  useEffect(() => {
    fetchChecklists();
  }, []);

  const fetchChecklists = async () => {
    const response = await fetch('/api/v1/checklists', {
      headers: { 'X-Client-Id': getClientId() },
      credentials: 'include',
    });
    const data = await response.json();
    setChecklists(data);
  };

  // Group checklists by ownership
  const ownedChecklists = checklists.filter(c => c.isOwner);
  const sharedChecklists = checklists.filter(c => !c.isOwner);

  return (
    <div>
      <section>
        <h2>My Checklists</h2>
        {ownedChecklists.map(checklist => (
          <ChecklistCard
            key={checklist.id}
            checklist={checklist}
            onShare={handleShare}
            {...otherProps}
          />
        ))}
      </section>

      {sharedChecklists.length > 0 && (
        <section>
          <h2>Shared with me</h2>
          {sharedChecklists.map(checklist => (
            <ChecklistCard
              key={checklist.id}
              checklist={checklist}
              {...otherProps}
            />
          ))}
        </section>
      )}
    </div>
  );
};
```

---

## Code Examples

### Utility Functions

```typescript
// Get or generate client ID (persisted per browser tab)
function getClientId(): string {
  let clientId = sessionStorage.getItem('clientId');
  if (!clientId) {
    clientId = generateUUID();
    sessionStorage.setItem('clientId', clientId);
  }
  return clientId;
}

// Generate UUID v4
function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

// API wrapper with auth and client ID
async function apiCall(
  endpoint: string,
  options: RequestInit = {}
): Promise<Response> {
  return fetch(endpoint, {
    ...options,
    headers: {
      'X-Client-Id': getClientId(),
      'Content-Type': 'application/json',
      ...options.headers,
    },
    credentials: 'include',
  });
}
```

### Toast Notifications

```typescript
// Simple toast notification hook
function useToast() {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const showToast = (message: string, type: 'success' | 'error' = 'success') => {
    const id = Date.now();
    setToasts([...toasts, { id, message, type }]);
    setTimeout(() => {
      setToasts(toasts => toasts.filter(t => t.id !== id));
    }, 3000);
  };

  return { toasts, showToast };
}

interface Toast {
  id: number;
  message: string;
  type: 'success' | 'error';
}
```

### Protected Route for Claim Page

```typescript
// Ensure user is authenticated before accessing claim page
const ProtectedClaimRoute: React.FC = () => {
  const { token } = useParams<{ token: string }>();
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    const returnUrl = encodeURIComponent(`/invites/${token}/claim`);
    return <Navigate to={`/login?returnUrl=${returnUrl}`} />;
  }

  return <ClaimInvitePage />;
};

// Route configuration
<Route path="/invites/:token/claim" element={<ProtectedClaimRoute />} />
```

---

## Error Handling

### HTTP Error Codes

| Code | Meaning | User-Facing Message | Action |
|------|---------|-------------------|--------|
| 200 | Success | "Successfully joined checklist" | Redirect to checklist |
| 201 | Created | "Invite link created" | Copy to clipboard |
| 204 | No Content | "Invite revoked" | Refresh invite list |
| 400 | Bad Request (Expired) | "This invite link has expired" | Show error state |
| 400 | Bad Request (Claimed) | "This invite has already been used" | Show error state |
| 400 | Bad Request (Invalid) | "Invalid invite settings" | Show error in form |
| 401 | Unauthorized | "Please log in to continue" | Redirect to login |
| 403 | Forbidden | "You don't have permission to manage invites" | Show error message |
| 404 | Not Found (Invite) | "Invalid invite link" | Show error state |
| 404 | Not Found (Checklist) | "Checklist not found" | Show error state |
| 500 | Server Error | "Something went wrong. Please try again." | Show error + retry button |

### Error Handling Patterns

```typescript
// Wrapper for API calls with error handling
async function handleApiError(response: Response): Promise<void> {
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Unknown error' }));

    switch (response.status) {
      case 401:
        // Redirect to login
        window.location.href = `/login?returnUrl=${encodeURIComponent(window.location.pathname)}`;
        break;
      case 403:
        showToast('You don\'t have permission to perform this action', 'error');
        break;
      case 404:
        showToast('Resource not found', 'error');
        break;
      case 500:
        showToast('Server error. Please try again later.', 'error');
        break;
      default:
        showToast(error.message || 'An error occurred', 'error');
    }

    throw new Error(error.message);
  }
}

// Usage
try {
  const response = await apiCall('/api/v1/checklists/123/invites', { method: 'POST', ... });
  await handleApiError(response);
  const data = await response.json();
  // Success handling
} catch (err) {
  // Error already handled by handleApiError
}
```

---

## Edge Cases

### 1. User Already Has Access

**Scenario:** User clicks an invite link for a checklist they already have access to.

**Backend Behavior:** API returns 200 with checklistId (idempotent)

**Frontend Handling:**
```typescript
// ClaimInvitePage handles this automatically
if (response.ok) {
  // Works the same whether it's first claim or already has access
  setStatus('success');
  navigate(`/checklists/${data.checklistId}`);
}
```

**User Experience:** User sees success message and is redirected to the checklist. No error is shown.

---

### 2. Invite Expires While User Is on Claim Page

**Scenario:** User opens invite link, but it expires before they click "Claim"

**Backend Behavior:** API returns 400 with "Invite expired" message

**Frontend Handling:**
```typescript
// Show clear error message
setStatus('error');
setErrorMessage('This invite link has expired. Please request a new invite from the checklist owner.');
```

**User Experience:** Clear error message with instruction to get new invite.

---

### 3. Network Error During Claim

**Scenario:** Network fails during the claim API call

**Frontend Handling:**
```typescript
try {
  const response = await apiCall(`/api/v1/invites/${token}/claim`, { method: 'POST' });
  // ...
} catch (err) {
  setStatus('error');
  setErrorMessage('Network error. Please check your connection and try again.');
  // Show retry button
  setShowRetryButton(true);
}
```

**User Experience:** Error message with retry button that re-attempts the claim.

---

### 4. Multiple Claims of Same Link (Race Condition)

**Scenario:** User opens link in two browser tabs and both try to claim simultaneously

**Backend Behavior:**
- Backend uses serializable transaction isolation
- First claim succeeds (200)
- Second claim also succeeds because it's idempotent (same user claiming again)

**Frontend Handling:** No special handling needed. Both tabs show success and redirect.

---

### 5. Owner Revokes Invite While Recipient Is Viewing Link

**Scenario:** Owner clicks "Revoke" on invite while recipient has the claim page open

**Backend Behavior:**
- Invite is deleted from database
- When recipient clicks claim: 404 "Invite not found"

**Frontend Handling:**
```typescript
if (response.status === 404) {
  setStatus('error');
  setErrorMessage('This invite link is no longer valid. It may have been revoked by the owner.');
}
```

**User Experience:** Clear message that invite was revoked.

---

### 6. Checklist Deleted After Invite Created

**Scenario:** Owner deletes checklist but invite still exists

**Backend Behavior:**
- Claim API checks if checklist exists
- Returns 404 "Checklist not found"

**Frontend Handling:**
```typescript
if (response.status === 404) {
  setStatus('error');
  setErrorMessage('This checklist no longer exists.');
}
```

---

### 7. User Logs Out and Logs Back In with Different Account

**Scenario:** User receives invite link while logged in as User A, logs out, logs in as User B, then claims invite

**Backend Behavior:** Invite is claimed by User B's Google ID

**Frontend Handling:** No special handling needed. Works normally.

**User Experience:** User B gains access to the checklist.

---

### 8. Clipboard API Not Available

**Scenario:** Browser doesn't support `navigator.clipboard` (e.g., non-HTTPS context)

**Frontend Handling:**
```typescript
async function copyToClipboard(text: string): Promise<boolean> {
  // Try modern clipboard API
  if (navigator.clipboard) {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch (err) {
      console.error('Clipboard API failed:', err);
    }
  }

  // Fallback for older browsers
  try {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);
    textarea.select();
    const success = document.execCommand('copy');
    document.body.removeChild(textarea);
    return success;
  } catch (err) {
    console.error('Fallback copy failed:', err);
    return false;
  }
}

// Usage
const success = await copyToClipboard(invite.inviteUrl);
if (success) {
  showToast('Link copied to clipboard!');
} else {
  // Show manual copy UI
  showManualCopyDialog(invite.inviteUrl);
}
```

**User Experience:** If auto-copy fails, show a dialog with the URL in a text field for manual copy.

---

### 9. Very Long Invite URLs

**Scenario:** Invite token is 64 characters, full URL could be long

**Frontend Handling:**
```typescript
// Truncate for display
function truncateUrl(url: string, maxLength: number = 40): string {
  if (url.length <= maxLength) return url;
  const start = url.substring(0, maxLength / 2);
  const end = url.substring(url.length - maxLength / 2);
  return `${start}...${end}`;
}

// In UI
<span title={invite.inviteUrl}>
  {truncateUrl(invite.inviteUrl)}
</span>
```

**User Experience:** URL is truncated in display but full URL is shown on hover and copied to clipboard.

---

### 10. User Spams "Generate Link" Button

**Scenario:** User clicks "Generate Link" button multiple times rapidly

**Frontend Handling:**
```typescript
const [isCreating, setIsCreating] = useState(false);

const handleGenerateLink = async () => {
  if (isCreating) return;  // Prevent duplicate requests

  setIsCreating(true);
  try {
    // API call...
  } finally {
    setIsCreating(false);
  }
};

// In render
<button onClick={handleGenerateLink} disabled={isCreating}>
  {isCreating ? 'Generating...' : 'Generate Link'}
</button>
```

**User Experience:** Button is disabled during creation to prevent spam.

---

## Testing Checklist

### Unit Tests

- [ ] `mapInviteToDisplay` correctly converts API response to display format
- [ ] `getExpiryLabel` returns correct labels for various expiry scenarios
- [ ] `getStatusLabel` returns correct status for active/claimed/expired
- [ ] `copyToClipboard` falls back to legacy method when clipboard API unavailable
- [ ] `generateUUID` generates valid UUID v4 format

### Integration Tests

- [ ] ShareChecklistModal fetches invites on open
- [ ] ShareChecklistModal creates invite and updates list
- [ ] ShareChecklistModal copies URL to clipboard on generate
- [ ] ShareChecklistModal revokes invite after confirmation
- [ ] ClaimInvitePage auto-claims on mount
- [ ] ClaimInvitePage redirects to login if not authenticated
- [ ] ClaimInvitePage redirects to checklist on success
- [ ] ClaimInvitePage shows error for expired invite
- [ ] ChecklistCard shows share icon for shared checklists
- [ ] ChecklistCard hides Share/Delete menu for non-owners
- [ ] ChecklistsPage groups checklists correctly

### E2E Tests

**Happy Path:**
- [ ] Owner creates checklist
- [ ] Owner opens share modal and generates invite link
- [ ] Invite URL is copied to clipboard
- [ ] Recipient (different user) opens invite URL
- [ ] Recipient is redirected to login if not authenticated
- [ ] After login, claim succeeds automatically
- [ ] Recipient is redirected to checklist detail page
- [ ] Checklist appears in recipient's "Shared with me" section
- [ ] Recipient can view and edit checklist items
- [ ] Real-time SSE updates work between owner and recipient

**Error Paths:**
- [ ] Expired invite shows error message
- [ ] Single-use invite claimed twice shows error on second claim
- [ ] Invalid invite token shows error
- [ ] Non-owner trying to create invite gets 403
- [ ] Revoking invite makes it unusable
- [ ] Deleted checklist invite shows error on claim

**Edge Cases:**
- [ ] User claiming invite they already have access to succeeds
- [ ] Multiple tabs claiming same invite both succeed
- [ ] Clipboard API fallback works in non-HTTPS context
- [ ] Long URLs are truncated in display
- [ ] Spamming "Generate Link" button doesn't create duplicate requests

### Manual Testing

- [ ] Share modal UI matches design
- [ ] Claim page loading state shows spinner
- [ ] Claim page success state shows checkmark
- [ ] Claim page error states show appropriate icons/messages
- [ ] Toast notifications appear and disappear correctly
- [ ] Invite list updates after generating new invite
- [ ] Invite list updates after revoking invite
- [ ] Shared checklist indicator (🔗) appears correctly
- [ ] Three-dot menu shows/hides options based on ownership
- [ ] Auto-redirect timing feels appropriate (not too fast/slow)

### Accessibility Testing

- [ ] Keyboard navigation works in share modal
- [ ] Focus trap works in modals
- [ ] Screen readers announce success/error states
- [ ] ARIA labels are present on interactive elements
- [ ] Color is not the only indicator of status (use icons too)
- [ ] Loading states are announced to screen readers

### Performance Testing

- [ ] Invite list loads quickly (< 500ms for 100 invites)
- [ ] Generating invite completes in < 1 second
- [ ] Claiming invite completes in < 1 second
- [ ] No memory leaks when opening/closing share modal repeatedly
- [ ] SSE connection doesn't break after claiming invite

---

## Privacy & Security Notes

### What NOT to Display to Users

**Never show Google IDs in the UI:**
- ❌ "Claimed by: 123456789012345678901"
- ✅ "Claimed on Jan 20, 2025 at 3:30 PM"

**For claimed invites, show:**
- Claim timestamp: "Claimed on Jan 20 at 3:30 PM"
- Generic label: "This invite has been claimed"
- Do NOT show: claimedBy field from API

**For shared users list (future feature):**
- ❌ "Shared with: user@gmail.com"
- ❌ "Shared with: 123456789012345678901"
- ✅ "3 users have access"
- ✅ "User joined via invite 2 days ago [Remove Access]"

### Security Best Practices

1. **HTTPS Only:** App should enforce HTTPS in production (backend already enforces this)

2. **Token Storage:** Never store invite tokens in localStorage (use sessionStorage if needed, but prefer URL-based flow)

3. **Auto-Claim:** Claim on page load prevents link sharing vulnerabilities (link can't be bookmarked and used later if single-use)

4. **Client ID:** Must be sent with every mutating request to prevent SSE echo

5. **Error Messages:** Don't leak information in error messages
   - ✅ "Invite not found" (could be invalid or revoked)
   - ❌ "Invite was revoked by user@gmail.com on [date]"

6. **Rate Limiting:** Backend should rate-limit invite creation and claim endpoints

---

## FAQ

### Q: Can I send the invite link via email from the app?

**A:** Not in the initial implementation. Users copy the link and share it via their preferred channel (WhatsApp, email, Slack, etc.). This maintains zero-PII constraint and gives users flexibility.

### Q: What happens if a user claims an invite for a checklist they already have access to?

**A:** The API is idempotent - it returns success (200) and the user is redirected to the checklist. No error is shown.

### Q: Can I see who has access to my checklist?

**A:** Not in the initial implementation. Future enhancement could show: "3 users have access" with anonymized entries like "User joined 2 days ago [Remove]".

### Q: Can I change invite settings after creation (e.g., extend expiration)?

**A:** No. To change settings, revoke the old invite and create a new one with different settings.

### Q: What's the maximum expiration time?

**A:** Backend accepts `expiresInHours` with max 8760 (1 year). Frontend offers: Never, 1 day, 7 days, 30 days.

### Q: Can I un-share a checklist after someone has claimed the invite?

**A:** Not in the initial implementation. Revoking an invite only prevents future claims. Removing existing access requires a separate feature (DELETE /api/v1/checklists/{id}/shares/{userId}).

### Q: Do invite links work if the checklist is deleted?

**A:** No. If the checklist is deleted, claiming the invite returns 404 "Checklist not found".

### Q: Can I share the same checklist with multiple people?

**A:** Yes. Create multiple invites (or one reusable invite) and share with multiple recipients. Each recipient claims the invite independently.

### Q: What happens to SSE connections when a new user claims an invite?

**A:** SSE connections work automatically. When a user claims an invite, they're added to the CHECKLIST_SHARE table. The guard rail on SSE subscribe checks this table, so the new user can connect and receive real-time updates.

### Q: Can recipients share the checklist with others?

**A:** No. Only the owner can create invites. Shared users cannot create invites or further share the checklist.

---

## Summary

**Implementation Checklist:**

1. ✅ Add ShareChecklistModal component
2. ✅ Add ClaimInvitePage component
3. ✅ Modify ChecklistCard to show shared indicator
4. ✅ Modify ChecklistsPage to group checklists
5. ✅ Add route: `/invites/:token/claim`
6. ✅ Implement clipboard copying with fallback
7. ✅ Add toast notification system
8. ✅ Add authentication check for claim route
9. ✅ Add error handling for all API calls
10. ✅ Test all user flows and edge cases

**Key Points:**

- Privacy-first: Never display Google IDs in UI
- Auto-claim: Claim happens on page load for better UX
- Idempotent: Claiming same invite twice succeeds
- Real-time: SSE works automatically for shared users
- Simple sharing: Copy link and share via any channel

**Questions or Clarifications:**

Contact the backend team if:
- API behavior differs from this document
- New error codes are added
- Additional fields are needed in responses
- Performance issues occur with large invite lists

---

**Document Version:** 1.0
**Last Updated:** 2025-01-21
**Maintained By:** Backend Team
**Feedback:** Create an issue in the repo or contact the team
