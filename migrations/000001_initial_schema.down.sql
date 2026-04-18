-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_user_addresses_updated_at ON user_addresses;
DROP TRIGGER IF EXISTS update_regions_updated_at ON regions;
DROP TRIGGER IF EXISTS update_region_configs_updated_at ON region_configs;
DROP TRIGGER IF EXISTS update_tax_categories_updated_at ON tax_categories;
DROP TRIGGER IF EXISTS update_tax_zones_updated_at ON tax_zones;
DROP TRIGGER IF EXISTS update_tax_rates_updated_at ON tax_rates;
DROP TRIGGER IF EXISTS update_restaurants_updated_at ON restaurants;
DROP TRIGGER IF EXISTS update_restaurant_hours_updated_at ON restaurant_hours;
DROP TRIGGER IF EXISTS update_menu_items_updated_at ON menu_items;
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS update_order_items_updated_at ON order_items;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS update_deliveries_updated_at ON deliveries;
DROP TRIGGER IF EXISTS update_couriers_updated_at ON couriers;
DROP TRIGGER IF EXISTS update_jobs_updated_at ON jobs;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column;

-- Drop indexes
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_entity;
DROP INDEX IF EXISTS idx_audit_logs_actor_id;

DROP INDEX IF EXISTS idx_idempotency_keys_expires_at;
DROP INDEX IF EXISTS idx_idempotency_keys_key_scope;

DROP INDEX IF EXISTS idx_jobs_priority;
DROP INDEX IF EXISTS idx_jobs_job_type;
DROP INDEX IF EXISTS idx_jobs_aggregate;
DROP INDEX IF EXISTS idx_jobs_status;

DROP INDEX IF EXISTS idx_outbox_events_aggregate;
DROP INDEX IF EXISTS idx_outbox_events_processed;

DROP INDEX IF EXISTS idx_couriers_location;
DROP INDEX IF EXISTS idx_couriers_is_available;
DROP INDEX IF EXISTS idx_couriers_user_id;

DROP INDEX IF EXISTS idx_deliveries_created_at;
DROP INDEX IF EXISTS idx_deliveries_status;
DROP INDEX IF EXISTS idx_deliveries_courier_id;
DROP INDEX IF EXISTS idx_deliveries_order_id;

DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_customer_id;
DROP INDEX IF EXISTS idx_payments_order_id;

DROP INDEX IF EXISTS idx_order_tax_lines_order_id;

DROP INDEX IF EXISTS idx_order_items_menu_item_id;
DROP INDEX IF EXISTS idx_order_items_order_id;

DROP INDEX IF EXISTS idx_orders_deleted_at;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_scheduled_for;
DROP INDEX IF EXISTS idx_orders_tax_zone_id;
DROP INDEX IF EXISTS idx_orders_region_id;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_restaurant_id;
DROP INDEX IF EXISTS idx_orders_customer_id;

DROP INDEX IF EXISTS idx_menu_items_deleted_at;
DROP INDEX IF EXISTS idx_menu_items_is_active;
DROP INDEX IF EXISTS idx_menu_items_is_available;
DROP INDEX IF EXISTS idx_menu_items_tax_category_id;
DROP INDEX IF EXISTS idx_menu_items_restaurant_id;

DROP INDEX IF EXISTS idx_restaurants_deleted_at;
DROP INDEX IF EXISTS idx_restaurants_location;
DROP INDEX IF EXISTS idx_restaurants_is_approved;
DROP INDEX IF EXISTS idx_restaurants_is_active;
DROP INDEX IF EXISTS idx_restaurants_region_id;
DROP INDEX IF EXISTS idx_restaurants_owner_id;

DROP INDEX IF EXISTS idx_tax_rates_deleted_at;
DROP INDEX IF EXISTS idx_tax_rates_is_active;
DROP INDEX IF EXISTS idx_tax_rates_effective_dates;
DROP INDEX IF EXISTS idx_tax_rates_tax_category_id;
DROP INDEX IF EXISTS idx_tax_rates_tax_zone_id;

DROP INDEX IF EXISTS idx_tax_zones_deleted_at;
DROP INDEX IF EXISTS idx_tax_zones_is_active;
DROP INDEX IF EXISTS idx_tax_zones_postal_pattern;
DROP INDEX IF EXISTS idx_tax_zones_city;
DROP INDEX IF EXISTS idx_tax_zones_state;
DROP INDEX IF EXISTS idx_tax_zones_country_code;

DROP INDEX IF EXISTS idx_regions_deleted_at;
DROP INDEX IF EXISTS idx_regions_is_active;
DROP INDEX IF EXISTS idx_regions_country_code;
DROP INDEX IF EXISTS idx_regions_code;

DROP INDEX IF EXISTS idx_user_addresses_deleted_at;
DROP INDEX IF EXISTS idx_user_addresses_user_id;

DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;

DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS couriers;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS order_tax_lines;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS menu_items;
DROP TABLE IF EXISTS restaurant_hours;
DROP TABLE IF EXISTS restaurants;
DROP TABLE IF EXISTS tax_rates;
DROP TABLE IF EXISTS tax_zones;
DROP TABLE IF EXISTS tax_categories;
DROP TABLE IF EXISTS region_configs;
DROP TABLE IF EXISTS regions;
DROP TABLE IF EXISTS user_addresses;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
