# Database Migrations - Safe Work with Existing Projects

## 🔒 Data Safety

The migration system is **completely safe** for existing projects:

- ✅ **Does not recreate existing data**
- ✅ **Applies only new migrations**
- ✅ **Tracks applied changes**
- ✅ **Supports incremental updates**

## 📋 Usage Scenarios

### 1. New Project (Empty Database)
```bash
# On first run, ALL migrations will be applied
go run cmd/main.go
# Result: All 9 migrations applied
```

### 2. Existing Project with Data
```bash
# On startup, only NEW migrations will be applied
go run cmd/main.go
# Result: Existing data preserved, new migrations applied
```

### 3. Project with Complete Migrations
```bash
# On startup, NO migrations will be applied
go run cmd/main.go
# Result: Project will start as usual
```

## 🛠️ Migration Management Utilities

### Check Migration Status
```bash
go run cmd/check-migrations/main.go
```

Shows:
- Which migrations are already applied
- Which tables exist
- Database status

### Example Output:
```
Applied migrations:
Migration applied version=20250615131409 applied_at=2025-10-24T07:48:55Z
Migration applied version=20250701000000 applied_at=2025-10-24T07:48:55Z
...
Total migrations applied count=9

Checking main tables:
Table exists table=users
Table exists table=products
Table exists table=launches
Table exists table=categories
```

## 🔄 How Migrations Work

### 1. Tracking
The system creates a `schema_migrations` table:
```sql
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 2. Startup Check
```go
// Gets list of applied migrations
appliedMigrations, err := getAppliedMigrations(db)

// Applies only NOT applied ones
for _, migration := range migrations {
    if _, applied := appliedMigrations[migration.Version]; !applied {
        // Apply migration
    }
}
```

### 3. Safe Application
- Each migration is applied in a transaction
- On error - rollback changes
- Record in `schema_migrations` only after successful application

## 📁 Migration Structure

```
migrations/
├── 20250615131409_state.sql      # Main tables + categories
├── 20250701000000_images.sql     # Image support
├── 20250702000000_user_bio.sql   # User biographies
├── 20250702010000_launch_url.sql # Launch URLs
├── 20250703000000_product_invitations.sql # Invitations
├── 20250704000000_launch_comments.sql # Comments
├── 20250705000000_awards.sql     # Awards
├── 20250706000000_drop_launch_slug_unique.sql # Fixes
└── 20250707000000_blog.sql       # Blog
```

## ⚠️ Important Points

### DO (Best Practices):
- ✅ Always backup before applying migrations in production
- ✅ Test migrations on data copies
- ✅ Use transactions in migrations
- ✅ Check status with `check-migrations`

### DON'T (Avoid):
- ❌ Don't delete migration files after applying
- ❌ Don't modify already applied migrations
- ❌ Don't apply migrations directly to database without tracking system

## 🚀 Adding New Migrations

1. **Create new file** in `migrations/` folder:
   ```
   20250708000000_new_feature.sql
   ```

2. **Use goose format**:
   ```sql
   -- +goose Up
   -- +goose StatementBegin
   
   CREATE TABLE new_table (
       id TEXT PRIMARY KEY,
       name VARCHAR(255) NOT NULL
   );
   
   -- +goose StatementEnd
   
   -- +goose Down
   -- +goose StatementBegin
   
   DROP TABLE new_table;
   
   -- +goose StatementEnd
   ```

3. **On next startup** migration will be applied automatically

## 🔍 Troubleshooting

### If migration doesn't apply:
```bash
# Check status
go run cmd/check-migrations/main.go

# Check startup logs
go run cmd/main.go
```

### If you need to rollback migration:
```sql
-- Remove record from schema_migrations
DELETE FROM schema_migrations WHERE version = 'VERSION_NUMBER';

-- Execute Down part of migration manually
```

## 📞 Support

When encountering migration issues:
1. Check application logs
2. Use `check-migrations` for diagnostics
3. Ensure database is not locked
4. Restore from backup if necessary
