#!/bin/bash
set -e

echo "==================================="
echo "ShopPilot Database Schema Validation"
echo "==================================="
echo ""

cd "$(dirname "$0")/.."
source settings/envVars.LOCAL

echo "✓ Environment loaded"
echo "  DATABASE_URL: ${DATABASE_URL}"
echo ""

# Check migration status
echo "📋 Migration Status:"
goose -dir migrations postgres "$DATABASE_URL" status
echo ""

# Validate Platform Schema
echo "🔍 Validating Platform Schema..."
echo "  - platform_users: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM platform_users;")"
echo "  - platform_permissions: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM platform_permissions;")"
echo "  - platform_roles: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM platform_roles;")"
echo "  - platform_role_permissions: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM platform_role_permissions;")"
echo ""

# Validate Client Schema
echo "🔍 Validating Client Schema..."
echo "  - clients: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM clients;")"
echo "  - client_users: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM client_users;")"
echo "  - client_permissions: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM client_permissions;")"
echo "  - client_roles: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM client_roles;")"
echo "  - client_role_permissions: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM client_role_permissions;")"
echo ""

# Validate Shop Schema
echo "🔍 Validating Shop Schema..."
echo "  - shops: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM shops;")"
echo "  - shop_users: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM shop_users;")"
echo "  - shops partitions:"
psql "$DATABASE_URL" -t -c "SELECT '    - ' || tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'shops_p%' ORDER BY tablename;"
echo ""

# Validate Product Schema
echo "🔍 Validating Product Schema..."
echo "  - products: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM products;")"
echo "  - product_variants: $(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM product_variants;")"
echo "  - products partitions:"
psql "$DATABASE_URL" -t -c "SELECT '    - ' || tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'products_p%' ORDER BY tablename;"
echo "  - product_variants partitions:"
psql "$DATABASE_URL" -t -c "SELECT '    - ' || tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'product_variants_p%' ORDER BY tablename;"
echo ""

# Validate Foreign Keys
echo "🔗 Validating Foreign Key Relationships..."
psql "$DATABASE_URL" -c "
SELECT
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
ORDER BY tc.table_name, kcu.column_name;
"
echo ""

# Validate Partitioning
echo "📊 Validating Partition Configuration..."
psql "$DATABASE_URL" -c "
SELECT
    parent.relname AS parent_table,
    child.relname AS partition_name,
    pg_get_expr(child.relpartbound, child.oid) AS partition_bounds
FROM pg_inherits
JOIN pg_class parent ON pg_inherits.inhparent = parent.oid
JOIN pg_class child ON pg_inherits.inhrelid = child.oid
WHERE parent.relname IN ('shops', 'products', 'product_variants')
ORDER BY parent.relname, child.relname;
"
echo ""

echo "✅ Schema validation complete!"
