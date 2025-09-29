# How to Append/Update API Keys

## Method 1: Firebase Console (Manual)

### Adding a New Key:
1. Go to Firebase Console → Firestore Database
2. Navigate to collection: `api_accesses`
3. Click on document: `document`
4. Find the `api_keys` array field
5. Click "Add item" in the array
6. Add the new key object:
```json
{
  "key": "ak_live_new1234567890abcdef1234567890",
  "expiry_time": "2025-12-31T23:59:59Z",
  "source": "manual",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "is_active": true
}
```
7. Update the document's `updated_at` field to current timestamp
8. Save the document

### Updating an Existing Key:
1. Find the key in the `api_keys` array
2. Click on the specific key object
3. Modify the fields (e.g., change `is_active` to `false`)
4. Update the `updated_at` field in both the key and document level
5. Save

## Method 2: Using Firebase Admin SDK (Node.js)

### Append a New Key:
```javascript
const admin = require('firebase-admin');

// Initialize Firebase Admin SDK
const serviceAccount = require('./path/to/serviceAccountKey.json');
admin.initializeApp({
  credential: admin.credential.cert(serviceAccount)
});

const db = admin.firestore();

async function appendAPIKey(newKey) {
  const docRef = db.collection('api_accesses').doc('document');
  
  try {
    await db.runTransaction(async (transaction) => {
      const doc = await transaction.get(docRef);
      
      if (!doc.exists) {
        // Create document if it doesn't exist
        const initialData = {
          api_keys: [newKey],
          updated_at: admin.firestore.Timestamp.now()
        };
        transaction.set(docRef, initialData);
      } else {
        const data = doc.data();
        data.api_keys.push(newKey);
        data.updated_at = admin.firestore.Timestamp.now();
        transaction.update(docRef, data);
      }
    });
    
    console.log('✅ API key appended successfully');
    return true;
  } catch (error) {
    console.error('❌ Error appending API key:', error);
    return false;
  }
}

async function updateAPIKey(keyToUpdate, updates) {
  const docRef = db.collection('api_accesses').doc('document');
  
  try {
    await db.runTransaction(async (transaction) => {
      const doc = await transaction.get(docRef);
      const data = doc.data();
      
      // Find and update the specific key
      const keyIndex = data.api_keys.findIndex(k => k.key === keyToUpdate);
      if (keyIndex === -1) {
        throw new Error(`API key ${keyToUpdate} not found`);
      }
      
      // Update the key
      data.api_keys[keyIndex] = {
        ...data.api_keys[keyIndex],
        ...updates,
        updated_at: admin.firestore.Timestamp.now()
      };
      
      data.updated_at = admin.firestore.Timestamp.now();
      transaction.update(docRef, data);
    });
    
    console.log('✅ API key updated successfully');
    return true;
  } catch (error) {
    console.error('❌ Error updating API key:', error);
    return false;
  }
}

// Example usage:

// 1. Add a new key
const newKey = {
  key: "ak_live_" + generateRandomString(32),
  expiry_time: admin.firestore.Timestamp.fromDate(new Date('2025-12-31T23:59:59Z')),
  source: "nodejs_script",
  created_at: admin.firestore.Timestamp.now(),
  updated_at: admin.firestore.Timestamp.now(),
  is_active: true
};

appendAPIKey(newKey);

// 2. Update existing key (deactivate)
updateAPIKey("ak_live_1234567890abcdef1234567890abcdef", {
  is_active: false
});

// 3. Update existing key (extend expiry)
updateAPIKey("ak_test_abcdef1234567890abcdef1234567890", {
  expiry_time: admin.firestore.Timestamp.fromDate(new Date('2026-12-31T23:59:59Z'))
});

function generateRandomString(length) {
  const chars = 'abcdef0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}
```

## Method 3: Using Go (Your API Context)

Create a utility script to manage keys:

```go
// goapi/cmd/manage-keys/main.go
package main

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "flag"
    "fmt"
    "log"
    "os"
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

func main() {
    var (
        action = flag.String("action", "", "Action: add, update, deactivate, list")
        key    = flag.String("key", "", "API key (for update/deactivate)")
        prefix = flag.String("prefix", "ak_live", "Prefix for new key")
        expiry = flag.String("expiry", "2025-12-31T23:59:59Z", "Expiry time")
        source = flag.String("source", "go_script", "Source of the key")
    )
    flag.Parse()

    if *action == "" {
        log.Fatal("Action is required: add, update, deactivate, list")
    }

    ctx := context.Background()
    
    // Initialize Firestore
    serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
    projectID := os.Getenv("FIREBASE_PROJECT_ID")
    
    client, err := firestore.NewClient(ctx, projectID, 
        option.WithCredentialsFile(serviceAccountPath))
    if err != nil {
        log.Fatalf("Failed to create Firestore client: %v", err)
    }
    defer client.Close()

    switch *action {
    case "add":
        err = addAPIKey(ctx, client, *prefix, *expiry, *source)
    case "update":
        if *key == "" {
            log.Fatal("Key is required for update action")
        }
        err = updateAPIKey(ctx, client, *key, map[string]interface{}{
            "expiry_time": parseTime(*expiry),
            "updated_at": time.Now().UTC(),
        })
    case "deactivate":
        if *key == "" {
            log.Fatal("Key is required for deactivate action")
        }
        err = updateAPIKey(ctx, client, *key, map[string]interface{}{
            "is_active": false,
            "updated_at": time.Now().UTC(),
        })
    case "list":
        err = listAPIKeys(ctx, client)
    default:
        log.Fatal("Unknown action. Use: add, update, deactivate, list")
    }

    if err != nil {
        log.Fatalf("Operation failed: %v", err)
    }
}

func addAPIKey(ctx context.Context, client *firestore.Client, prefix, expiry, source string) error {
    docRef := client.Collection("api_accesses").Doc("document")
    
    newKey := APIKey{
        Key:        generateAPIKey(prefix),
        ExpiryTime: parseTime(expiry),
        Source:     source,
        CreatedAt:  time.Now().UTC(),
        UpdatedAt:  time.Now().UTC(),
        IsActive:   true,
    }

    return client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
        doc, err := tx.Get(docRef)
        if err != nil {
            // Document doesn't exist, create it
            apiAccess := APIAccess{
                APIKeys:   []APIKey{newKey},
                UpdatedAt: time.Now().UTC(),
            }
            return tx.Set(docRef, apiAccess)
        }

        var apiAccess APIAccess
        if err := doc.DataTo(&apiAccess); err != nil {
            return fmt.Errorf("failed to parse document: %w", err)
        }

        apiAccess.APIKeys = append(apiAccess.APIKeys, newKey)
        apiAccess.UpdatedAt = time.Now().UTC()

        fmt.Printf("✅ Added new API key: %s\n", newKey.Key[:12]+"...")
        return tx.Set(docRef, apiAccess)
    })
}

func updateAPIKey(ctx context.Context, client *firestore.Client, keyToUpdate string, updates map[string]interface{}) error {
    docRef := client.Collection("api_accesses").Doc("document")
    
    return client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
        doc, err := tx.Get(docRef)
        if err != nil {
            return fmt.Errorf("failed to get document: %w", err)
        }

        var apiAccess APIAccess
        if err := doc.DataTo(&apiAccess); err != nil {
            return fmt.Errorf("failed to parse document: %w", err)
        }

        // Find the key to update
        keyIndex := -1
        for i, key := range apiAccess.APIKeys {
            if key.Key == keyToUpdate {
                keyIndex = i
                break
            }
        }

        if keyIndex == -1 {
            return fmt.Errorf("API key not found: %s", keyToUpdate)
        }

        // Apply updates
        if expiry, ok := updates["expiry_time"].(time.Time); ok {
            apiAccess.APIKeys[keyIndex].ExpiryTime = expiry
        }
        if active, ok := updates["is_active"].(bool); ok {
            apiAccess.APIKeys[keyIndex].IsActive = active
        }
        if updatedAt, ok := updates["updated_at"].(time.Time); ok {
            apiAccess.APIKeys[keyIndex].UpdatedAt = updatedAt
        }

        apiAccess.UpdatedAt = time.Now().UTC()

        fmt.Printf("✅ Updated API key: %s\n", keyToUpdate[:12]+"...")
        return tx.Set(docRef, apiAccess)
    })
}

func listAPIKeys(ctx context.Context, client *firestore.Client) error {
    doc, err := client.Collection("api_accesses").Doc("document").Get(ctx)
    if err != nil {
        return fmt.Errorf("failed to get document: %w", err)
    }

    var apiAccess APIAccess
    if err := doc.DataTo(&apiAccess); err != nil {
        return fmt.Errorf("failed to parse document: %w", err)
    }

    fmt.Printf("📋 Found %d API keys:\n\n", len(apiAccess.APIKeys))
    
    for i, key := range apiAccess.APIKeys {
        status := "🟢 Active"
        if !key.IsActive {
            status = "🔴 Inactive"
        }
        if time.Now().UTC().After(key.ExpiryTime) {
            status = "⏰ Expired"
        }

        fmt.Printf("%d. Key: %s... | %s | Expires: %s | Source: %s\n",
            i+1, key.Key[:12], status, key.ExpiryTime.Format("2006-01-02"), key.Source)
    }

    return nil
}

func generateAPIKey(prefix string) string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return prefix + "_" + hex.EncodeToString(bytes)
}

func parseTime(timeStr string) time.Time {
    t, err := time.Parse(time.RFC3339, timeStr)
    if err != nil {
        log.Fatalf("Invalid time format: %s", timeStr)
    }
    return t.UTC()
}
```

## Method 4: Using Your API Endpoints

### Refresh Cache After Manual Changes:
```bash
# After making changes in Firebase Console, refresh the cache
curl -X POST \
     -H "X-API-KEY: your-existing-key" \
     http://localhost:8080/api-keys/refresh-cache
```

### Clear Specific Key Cache:
```bash
# If you updated a specific key, clear its cache
curl -X DELETE \
     -H "X-API-KEY: your-admin-key" \
     http://localhost:8080/api-keys/ak_live_1234567890abcdef1234567890abcdef/cache
```

## Method 5: Python Script

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

def append_api_key(new_key_data):
    """Append a new API key to the document"""
    doc_ref = db.collection('api_accesses').document('document')
    
    @firestore.transactional
    def update_in_transaction(transaction):
        doc = doc_ref.get(transaction=transaction)
        
        if doc.exists:
            data = doc.to_dict()
        else:
            data = {'api_keys': []}
        
        # Append new key
        data['api_keys'].append(new_key_data)
        data['updated_at'] = datetime.now(timezone.utc)
        
        transaction.set(doc_ref, data)
        return len(data['api_keys'])
    
    transaction = db.transaction()
    count = update_in_transaction(transaction)
    print(f"✅ Added new key. Total keys: {count}")

def update_api_key(key_to_update, updates):
    """Update an existing API key"""
    doc_ref = db.collection('api_accesses').document('document')
    
    @firestore.transactional
    def update_in_transaction(transaction):
        doc = doc_ref.get(transaction=transaction)
        data = doc.to_dict()
        
        # Find and update the key
        for i, key_data in enumerate(data['api_keys']):
            if key_data['key'] == key_to_update:
                data['api_keys'][i].update(updates)
                data['api_keys'][i]['updated_at'] = datetime.now(timezone.utc)
                data['updated_at'] = datetime.now(timezone.utc)
                transaction.set(doc_ref, data)
                return True
        
        raise ValueError(f"Key {key_to_update} not found")
    
    transaction = db.transaction()
    update_in_transaction(transaction)
    print(f"✅ Updated key: {key_to_update[:12]}...")

# Example usage:
if __name__ == "__main__":
    # Add a new key
    new_key = {
        'key': f"ak_live_{''.join(secrets.choice(string.ascii_lowercase + string.digits) for _ in range(32))}",
        'expiry_time': datetime(2025, 12, 31, 23, 59, 59, tzinfo=timezone.utc),
        'source': 'python_script',
        'created_at': datetime.now(timezone.utc),
        'updated_at': datetime.now(timezone.utc),
        'is_active': True
    }
    
    append_api_key(new_key)
    
    # Update existing key (deactivate)
    update_api_key("ak_live_1234567890abcdef1234567890abcdef", {
        'is_active': False
    })
```

## Usage Examples:

### Using the Go script:
```bash
# Add a new key
go run goapi/cmd/manage-keys/main.go -action=add -prefix=ak_prod -source=manual

# Deactivate a key  
go run goapi/cmd/manage-keys/main.go -action=deactivate -key=ak_live_1234567890abcdef1234567890abcdef

# List all keys
go run goapi/cmd/manage-keys/main.go -action=list

# Update expiry
go run goapi/cmd/manage-keys/main.go -action=update -key=ak_test_abcdef1234567890abcdef1234567890 -expiry=2026-12-31T23:59:59Z
```

## Important Notes:

1. **Always use transactions** when updating to avoid race conditions
2. **Update the document-level `updated_at`** field when making changes
3. **Refresh the cache** after manual changes using the API endpoint
4. **Test your changes** with the API endpoints before deploying
5. **Keep backups** of your API keys document before major changes

Choose the method that best fits your workflow and technical requirements!
