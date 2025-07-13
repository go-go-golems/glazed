#!/bin/bash

# Run all layer guide tests individually
# Since they all have main functions, they need to be run separately

echo "Running all Glazed Layers Guide test programs..."
echo "================================================"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../../../"  # Go to glazed root

FAILED=0
PASSED=0

for i in {01..08}; do
    TEST_FILE="test/tutorials/layers-guide/${i}_*.go"
    TEST_NAME=$(ls $TEST_FILE 2>/dev/null | head -1 | xargs basename)
    
    if [ -z "$TEST_NAME" ]; then
        echo "❌ Test $i: File not found"
        FAILED=$((FAILED + 1))
        continue
    fi
    
    echo ""
    echo "Running Test $i: $TEST_NAME"
    echo "----------------------------------------"
    
    if go run $TEST_FILE; then
        echo "✅ Test $i: PASSED"
        PASSED=$((PASSED + 1))
    else
        echo "❌ Test $i: FAILED"
        FAILED=$((FAILED + 1))
    fi
done

echo ""
echo "================================================"
echo "Test Results Summary:"
echo "  Passed: $PASSED"
echo "  Failed: $FAILED"
echo "  Total:  $((PASSED + FAILED))"

if [ $FAILED -eq 0 ]; then
    echo "🎉 All tests passed!"
    exit 0
else
    echo "💥 Some tests failed!"
    exit 1
fi
