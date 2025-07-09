#!/bin/bash

# Script to switch between Jellyfin and Emby databases for development

set -e

EMBY_DB="./emby.db"
JELLYFIN_DB="./jellyfin.db"
CONFIG_FILE="backend/config/config.dev.yaml"

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Config file not found at $CONFIG_FILE"
    echo "Please create the config file first"
    exit 1
fi

run_migrations() {
    local db_file="$1"
    local db_name="$2"
    
    echo "🗄️  Running migrations for $db_name database..."
    cd backend
    
    # Run Atlas migrations
    if command -v atlas >/dev/null 2>&1; then
        atlas migrate apply \
            --url "sqlite://$(pwd)/$db_file" \
            --dir "file://migrations"
        echo "✅ Migrations completed for $db_name"
    else
        echo "❌ Atlas CLI not found. Please install Atlas to run migrations automatically."
        echo "💡 Or run manually: cd backend && make migrate"
        return 1
    fi
    
    cd ..
}

case "$1" in
    "jellyfin")
        echo "🔄 Switching to Jellyfin database..."
        
        # Check if database exists, run migrations if not
        if [ ! -f "backend/jellyfin.db" ]; then
            echo "📦 Jellyfin database doesn't exist, creating and running migrations..."
            run_migrations "jellyfin.db" "Jellyfin"
        fi
        
        # Update config file to use jellyfin.db
        if grep -q "path:" "$CONFIG_FILE"; then
            sed -i 's|path: .*|path: '"$JELLYFIN_DB"'|' "$CONFIG_FILE"
        else
            echo "❌ Could not find 'path:' in config file"
            exit 1
        fi
        
        # Update media server settings if they exist
        if grep -q "url:" "$CONFIG_FILE"; then
            sed -i 's|url: .*|url: http://localhost:8096|' "$CONFIG_FILE"
        fi
        if grep -q "type:" "$CONFIG_FILE"; then
            sed -i 's|type: .*|type: jellyfin|' "$CONFIG_FILE"
        fi
        
        echo "✅ Switched to Jellyfin database ($JELLYFIN_DB)"
        echo "💡 Configure Jellyfin at http://localhost:8096"
        echo "🔧 Restart the backend to use the new database"
        ;;
    "emby")
        echo "🔄 Switching to Emby database..."
        
        # Check if database exists, run migrations if not
        if [ ! -f "backend/emby.db" ]; then
            echo "📦 Emby database doesn't exist, creating and running migrations..."
            run_migrations "emby.db" "Emby"
        fi
        
        # Update config file to use emby.db
        if grep -q "path:" "$CONFIG_FILE"; then
            sed -i 's|path: .*|path: '"$EMBY_DB"'|' "$CONFIG_FILE"
        else
            echo "❌ Could not find 'path:' in config file"
            exit 1
        fi
        
        # Update media server settings if they exist
        if grep -q "url:" "$CONFIG_FILE"; then
            sed -i 's|url: .*|url: http://localhost:8097|' "$CONFIG_FILE"
        fi
        if grep -q "type:" "$CONFIG_FILE"; then
            sed -i 's|type: .*|type: emby|' "$CONFIG_FILE"
        fi
        
        echo "✅ Switched to Emby database ($EMBY_DB)"
        echo "💡 Configure Emby at http://localhost:8097"
        echo "🔧 Restart the backend to use the new database"
        ;;
    "status")
        echo "📊 Current Database Configuration:"
        echo ""
        
        if [ -f "$CONFIG_FILE" ]; then
            echo "Config file: $CONFIG_FILE"
            if grep -q "path:" "$CONFIG_FILE"; then
                db_path=$(grep "path:" "$CONFIG_FILE" | sed 's/.*path: *//')
                echo "Database path: $db_path"
                
                case "$db_path" in
                    *"jellyfin.db")
                        echo "Currently configured for: Jellyfin"
                        ;;
                    *"emby.db")
                        echo "Currently configured for: Emby"
                        ;;
                    *)
                        echo "Currently configured for: Unknown/Custom ($db_path)"
                        ;;
                esac
            else
                echo "No database path found in config"
            fi
            
            # Show media server settings if they exist
            if grep -q "url:" "$CONFIG_FILE"; then
                url=$(grep "url:" "$CONFIG_FILE" | sed 's/.*url: *//')
                echo "Media server URL: $url"
            fi
            if grep -q "type:" "$CONFIG_FILE"; then
                type=$(grep "type:" "$CONFIG_FILE" | sed 's/.*type: *//')
                echo "Media server type: $type"
            fi
        else
            echo "Config file not found: $CONFIG_FILE"
        fi
        
        echo ""
        echo "Available databases:"
        [ -f "backend/emby.db" ] && echo "  - Emby: ✅" || echo "  - Emby: ❌"
        [ -f "backend/jellyfin.db" ] && echo "  - Jellyfin: ✅" || echo "  - Jellyfin: ❌"
        ;;
    *)
        echo "Usage: $0 {jellyfin|emby|status}"
        echo ""
        echo "Examples:"
        echo "  $0 jellyfin  # Switch to Jellyfin database"
        echo "  $0 emby      # Switch to Emby database" 
        echo "  $0 status    # Show current database status"
        echo ""
        echo "Note: This modifies $CONFIG_FILE and requires backend restart."
        exit 1
        ;;
esac