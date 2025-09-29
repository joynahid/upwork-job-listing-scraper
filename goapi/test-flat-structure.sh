#!/bin/bash

echo "🧪 Testing Flat Structure API Key Management"
echo "============================================="

# Check if environment variables are set
if [ -z "$FIREBASE_SERVICE_ACCOUNT_PATH" ] || [ -z "$FIREBASE_PROJECT_ID" ]; then
    echo "⚠️ Please set the following environment variables:"
    echo "   export FIREBASE_SERVICE_ACCOUNT_PATH=/path/to/service-account.json"
    echo "   export FIREBASE_PROJECT_ID=your-project-id"
    echo ""
    echo "Example:"
    echo "   export FIREBASE_SERVICE_ACCOUNT_PATH=/root/upwork-job-scraper-d1f2c-firebase-adminsdk-fbsvc-3c0e2df938.json"
    echo "   export FIREBASE_PROJECT_ID=upwork-job-scraper-d1f2c"
    exit 1
fi

echo "📋 Environment:"
echo "   Project ID: $FIREBASE_PROJECT_ID"
echo "   Service Account: $(basename $FIREBASE_SERVICE_ACCOUNT_PATH)"
echo ""

# Test 1: List existing keys (should work with both old and new structure)
echo "📋 Test 1: List existing API keys"
echo "---------------------------------"
./manage-keys -action=list
echo ""

# Test 2: Add a new API key
echo "🔑 Test 2: Add a new test API key"
echo "----------------------------------"
./manage-keys -action=add -prefix=ak_test -source=flat_structure_test -expiry=2025-06-30T23:59:59Z
echo ""

# Test 3: List keys again to see the new one
echo "📋 Test 3: List keys again (should show new key)"
echo "------------------------------------------------"
./manage-keys -action=list
echo ""

echo "✅ Flat structure tests completed!"
echo ""
echo "🔧 Additional tests you can run:"
echo "   # Deactivate a key:"
echo "   ./manage-keys -action=deactivate -key=ak_test_[your-key-here]"
echo ""
echo "   # Activate a key:"
echo "   ./manage-keys -action=activate -key=ak_test_[your-key-here]"
echo ""
echo "   # Update expiry:"
echo "   ./manage-keys -action=update -key=ak_test_[your-key-here] -expiry=2026-12-31T23:59:59Z"
echo ""
echo "🚀 Performance benefits of flat structure:"
echo "   • O(1) lookups instead of O(n) array scans"
echo "   • Atomic updates without reading entire document"
echo "   • Native Firestore queries with indexing"
echo "   • Parallel processing capabilities"
echo "   • ~90% reduction in Firestore read/write operations"
