# API Key Creation Examples

## Firestore Document Structure

The API keys are stored in Firestore at the path: `api_accesses/document`

### Complete Document Structure
```json
{
  "api_keys": [
    {
      "key": "ak_live_1234567890abcdef1234567890abcdef",
      "expiry_time": "2025-12-31T23:59:59Z",
      "source": "dashboard",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "is_active": true
    }
  ],
  "updated_at": "2024-06-01T00:00:00Z"
}
```

## Adding a New API Key

### Method 1: Using Firebase Console
1. Go to your Firebase Console
2. Navigate to Firestore Database
3. Find collection: `api_accesses`
4. Open document: `document`
5. Add new object to the `api_keys` array

### Method 2: Using Firebase Admin SDK (Node.js)
```javascript
const admin = require('firebase-admin');

// Initialize Firebase Admin SDK
const serviceAccount = require('./path/to/serviceAccountKey.json');
admin.initializeApp({
  credential: admin.credential.cert(serviceAccount)
});

const db = admin.firestore();

async function addAPIKey(newAPIKey) {
  const docRef = db.collection('api_accesses').doc('document');
  
  try {
    const doc = await docRef.get();
    const data = doc.data();
    
    // Add new API key to existing array
    data.api_keys.push(newAPIKey);
    data.updated_at = new Date().toISOString();
    
    await docRef.set(data);
    console.log('API key added successfully');
  } catch (error) {
    console.error('Error adding API key:', error);
  }
}

// Example usage
const newKey = {
  key: "ak_prod_" + generateRandomString(32),
  expiry_time: "2025-12-31T23:59:59Z",
  source: "script",
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  is_active: true
};

addAPIKey(newKey);

function generateRandomString(length) {
  const chars = 'abcdef0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}
```

### Method 3: Using Go (Firebase Admin SDK)
```go
package main

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"

    "cloud.google.com/go/firestore"
    "google.golang.org/api/option"
)

type APIKey struct {
    Key        string    `json:"key" firestore:"key"`
    ExpiryTime time.Time `json:"expiry_time" firestore:"expiry_time"`
    Source     string    `json:"source" firestore:"source"`
    CreatedAt  time.Time `json:"created_at" firestore:"created_at"`
    UpdatedAt  time.Time `json:"updated_at" firestore:"updated_at"`
    IsActive   bool      `json:"is_active" firestore:"is_active"`
}

type APIAccess struct {
    APIKeys   []APIKey  `json:"api_keys" firestore:"api_keys"`
    UpdatedAt time.Time `json:"updated_at" firestore:"updated_at"`
}

func addAPIKey(ctx context.Context, client *firestore.Client, newKey APIKey) error {
    docRef := client.Collection("api_accesses").Doc("document")
    
    doc, err := docRef.Get(ctx)
    if err != nil {
        return fmt.Errorf("failed to get document: %w", err)
    }
    
    var apiAccess APIAccess
    if err := doc.DataTo(&apiAccess); err != nil {
        return fmt.Errorf("failed to parse document: %w", err)
    }
    
    // Add new key
    apiAccess.APIKeys = append(apiAccess.APIKeys, newKey)
    apiAccess.UpdatedAt = time.Now().UTC()
    
    _, err = docRef.Set(ctx, apiAccess)
    if err != nil {
        return fmt.Errorf("failed to update document: %w", err)
    }
    
    return nil
}

func generateAPIKey(prefix string) string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return prefix + "_" + hex.EncodeToString(bytes)
}

func main() {
    ctx := context.Background()
    
    // Initialize Firestore client
    client, err := firestore.NewClient(ctx, "your-project-id", 
        option.WithCredentialsFile("path/to/service-account.json"))
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    // Create new API key
    newKey := APIKey{
        Key:        generateAPIKey("ak_prod"),
        ExpiryTime: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
        Source:     "go_script",
        CreatedAt:  time.Now().UTC(),
        UpdatedAt:  time.Now().UTC(),
        IsActive:   true,
    }
    
    if err := addAPIKey(ctx, client, newKey); err != nil {
        fmt.Printf("Error adding API key: %v\n", err)
    } else {
        fmt.Printf("Successfully added API key: %s\n", newKey.Key[:12]+"...")
    }
}
```

### Method 4: Using Python (Firebase Admin SDK)
```python
import firebase_admin
from firebase_admin import credentials, firestore
from datetime import datetime, timezone
import secrets
import string

# Initialize Firebase Admin SDK
cred = credentials.Certificate('path/to/serviceAccountKey.json')
firebase_admin.initialize_app(cred)
db = firestore.client()

def generate_api_key(prefix="ak_prod", length=32):
    """Generate a secure random API key"""
    alphabet = string.ascii_lowercase + string.digits
    random_part = ''.join(secrets.choice(alphabet) for _ in range(length))
    return f"{prefix}_{random_part}"

def add_api_key(new_key_data):
    """Add a new API key to Firestore"""
    doc_ref = db.collection('api_accesses').document('document')
    
    try:
        doc = doc_ref.get()
        if doc.exists:
            data = doc.to_dict()
        else:
            data = {'api_keys': [], 'updated_at': datetime.now(timezone.utc)}
        
        # Add new API key
        data['api_keys'].append(new_key_data)
        data['updated_at'] = datetime.now(timezone.utc)
        
        doc_ref.set(data)
        print(f"Successfully added API key: {new_key_data['key'][:12]}...")
        
    except Exception as e:
        print(f"Error adding API key: {e}")

# Example usage
new_key = {
    'key': generate_api_key('ak_prod'),
    'expiry_time': datetime(2025, 12, 31, 23, 59, 59, tzinfo=timezone.utc),
    'source': 'python_script',
    'created_at': datetime.now(timezone.utc),
    'updated_at': datetime.now(timezone.utc),
    'is_active': True
}

add_api_key(new_key)
```

## API Key Naming Convention

Recommended naming pattern: `{prefix}_{environment}_{random_string}`

### Prefixes:
- `ak_` - API Key
- `tk_` - Test Key  
- `sk_` - Service Key

### Environments:
- `live` - Production
- `test` - Testing
- `dev` - Development
- `staging` - Staging

### Examples:
- `ak_live_1234567890abcdef1234567890abcdef`
- `ak_test_abcdef1234567890abcdef1234567890`
- `ak_dev_fedcba0987654321fedcba0987654321`

## Field Descriptions

- **key**: The actual API key string (32-64 characters recommended)
- **expiry_time**: ISO 8601 timestamp when key expires
- **source**: Where/how the key was created (dashboard, api, script, etc.)
- **created_at**: When the key was first created
- **updated_at**: When the key was last modified
- **is_active**: Boolean flag to enable/disable key instantly

## Testing Your New API Key

```bash
# Test the API key
curl -H "X-API-KEY: ak_live_1234567890abcdef1234567890abcdef" \
     http://localhost:8080/health

# Expected response
{
  "success": true,
  "message": "API is healthy",
  "last_updated": "2024-01-01T12:00:00Z"
}
```

## Managing API Keys

### Deactivate a key (without deleting):
```json
{
  "key": "ak_live_1234567890abcdef1234567890abcdef",
  "is_active": false,
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Refresh cache after changes:
```bash
curl -X POST \
     -H "X-API-KEY: your-admin-key" \
     http://localhost:8080/api-keys/refresh-cache
```
