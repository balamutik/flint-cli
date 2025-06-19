#!/bin/bash

# Resource monitoring script for Flint Vault stress test
OPERATION="$1"
VAULT_PID="$2"

if [ -z "$OPERATION" ]; then
    echo "Usage: $0 <operation_name> [vault_pid]"
    exit 1
fi

echo "🔍 MONITORING: $OPERATION"
echo "📊 Timestamp: $(date)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Monitor system resources
echo "💻 SYSTEM RESOURCES:"
echo "   Memory Usage:"
free -h | grep -E "Mem:|Swap:"

echo ""
echo "   CPU Load:"
uptime

echo ""
echo "   Disk Space:"
df -h . | tail -1

if [ ! -z "$VAULT_PID" ]; then
    echo ""
    echo "🔥 PROCESS SPECIFIC (PID: $VAULT_PID):"
    if ps -p $VAULT_PID > /dev/null 2>&1; then
        echo "   Memory (RSS/VSZ):"
        ps -o pid,rss,vsz,pmem,pcpu,cmd -p $VAULT_PID | tail -1
        
        echo ""
        echo "   Memory Details:"
        cat /proc/$VAULT_PID/status | grep -E "VmRSS|VmSize|VmPeak|VmHWM"
    else
        echo "   Process $VAULT_PID not found or already finished"
    fi
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "" 