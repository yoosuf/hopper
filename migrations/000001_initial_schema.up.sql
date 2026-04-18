-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Core Tables

-- Idempotency keys table
CREATE TABLE idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    scope VARCHAR(100) NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    response_code INTEGER NOT NULL,
    response_body JSONB,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX idx_idempotency_keys_scope ON idempotency_keys(scope);
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);

-- Reviews table
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(50) NOT NULL CHECK (target_type IN ('restaurant', 'courier', 'menu_item')),
    target_id UUID NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX idx_reviews_user_id ON reviews(user_id);
CREATE INDEX idx_reviews_target ON reviews(target_type, target_id);
CREATE INDEX idx_reviews_deleted_at ON reviews(deleted_at);

-- Promotions table
CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('coupon', 'discount', 'free_delivery')),
    discount_type VARCHAR(50) NOT NULL CHECK (discount_type IN ('percentage', 'fixed')),
    discount_value INTEGER NOT NULL CHECK (discount_value > 0),
    min_order_value INTEGER NOT NULL DEFAULT 0,
    max_discount INTEGER NOT NULL DEFAULT 0,
    usage_limit INTEGER NOT NULL DEFAULT 0,
    usage_count INTEGER NOT NULL DEFAULT 0,
    region_id UUID REFERENCES regions(id) ON DELETE CASCADE,
    restaurant_id UUID REFERENCES restaurants(id) ON DELETE CASCADE,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX idx_promotions_code ON promotions(code);
CREATE INDEX idx_promotions_region_id ON promotions(region_id);
CREATE INDEX idx_promotions_restaurant_id ON promotions(restaurant_id);
CREATE INDEX idx_promotions_active ON promotions(is_active);
CREATE INDEX idx_promotions_validity ON promotions(valid_from, valid_until);
CREATE INDEX idx_promotions_deleted_at ON promotions(deleted_at);

-- Notification preferences table
CREATE TABLE notification_preferences (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(50) NOT NULL CHECK (channel IN ('push', 'email', 'sms')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    order_events BOOLEAN NOT NULL DEFAULT true,
    payment_events BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, channel)
);

CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);

-- Support tickets table
CREATE TABLE support_tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('open', 'in_progress', 'resolved', 'closed')) DEFAULT 'open',
    priority VARCHAR(50) NOT NULL CHECK (priority IN ('low', 'medium', 'high', 'urgent')) DEFAULT 'medium',
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_support_tickets_user_id ON support_tickets(user_id);
CREATE INDEX idx_support_tickets_status ON support_tickets(status);
CREATE INDEX idx_support_tickets_assigned_to ON support_tickets(assigned_to);
CREATE INDEX idx_support_tickets_deleted_at ON support_tickets(deleted_at);

-- Support messages table
CREATE TABLE support_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_support BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX idx_support_messages_ticket_id ON support_messages(ticket_id);
CREATE INDEX idx_support_messages_user_id ON support_messages(user_id);

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('customer', 'restaurant_owner', 'courier', 'admin')),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    is_suspended BOOLEAN NOT NULL DEFAULT false,
    suspended_at TIMESTAMP WITH TIME ZONE,
    suspended_reason TEXT,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Refresh tokens table
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- User addresses table
CREATE TABLE user_addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    street_address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    state_or_province VARCHAR(100),
    postal_code VARCHAR(20) NOT NULL,
    country_code VARCHAR(2) NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    is_default BOOLEAN NOT NULL DEFAULT false,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Regions table
CREATE TABLE regions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    country_code VARCHAR(2) NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    currency_code VARCHAR(3) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Region configs table
CREATE TABLE region_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    region_id UUID NOT NULL REFERENCES regions(id) ON DELETE CASCADE,
    platform_fee_basis_points INTEGER NOT NULL DEFAULT 0 CHECK (platform_fee_basis_points >= 0),
    default_delivery_window_minutes INTEGER NOT NULL DEFAULT 30 CHECK (default_delivery_window_minutes > 0),
    order_activation_lead_minutes INTEGER NOT NULL DEFAULT 15 CHECK (order_activation_lead_minutes >= 0),
    allow_scheduled_orders BOOLEAN NOT NULL DEFAULT true,
    delivery_fee_taxable_default BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    UNIQUE (region_id)
);

-- Tax categories table
CREATE TABLE tax_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Tax zones table
CREATE TABLE tax_zones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    country_code VARCHAR(2) NOT NULL,
    state_or_province VARCHAR(100),
    city VARCHAR(100),
    postal_code_pattern VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT true,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Tax rates table
CREATE TABLE tax_rates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tax_zone_id UUID NOT NULL REFERENCES tax_zones(id) ON DELETE CASCADE,
    tax_category_id UUID NOT NULL REFERENCES tax_categories(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    rate_bps INTEGER NOT NULL CHECK (rate_bps >= 0),
    is_inclusive BOOLEAN NOT NULL DEFAULT false,
    applies_to_delivery_fee BOOLEAN NOT NULL DEFAULT false,
    effective_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    effective_to TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Restaurants table
CREATE TABLE restaurants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    cuisine_type VARCHAR(100),
    street_address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    state_or_province VARCHAR(100),
    postal_code VARCHAR(20) NOT NULL,
    country_code VARCHAR(2) NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    phone VARCHAR(20),
    email VARCHAR(255),
    region_id UUID NOT NULL REFERENCES regions(id) ON DELETE CASCADE,
    currency_code VARCHAR(3) NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_approved BOOLEAN NOT NULL DEFAULT false,
    approved_at TIMESTAMP WITH TIME ZONE,
    approved_by_id UUID REFERENCES users(id),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Restaurant hours table
CREATE TABLE restaurant_hours (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    is_closed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    UNIQUE (restaurant_id, day_of_week)
);

-- Menu items table
CREATE TABLE menu_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price_amount INTEGER NOT NULL CHECK (price_amount >= 0),
    currency_code VARCHAR(3) NOT NULL,
    tax_category_id UUID REFERENCES tax_categories(id) ON DELETE SET NULL,
    is_available BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Orders table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending_payment', 'scheduled', 'pending', 'accepted', 'rejected', 'preparing', 'ready_for_pickup', 'picked_up', 'delivered', 'cancelled')),
    delivery_type VARCHAR(20) NOT NULL CHECK (delivery_type IN ('asap', 'scheduled')),
    scheduled_for TIMESTAMP WITH TIME ZONE,
    delivery_window_start TIMESTAMP WITH TIME ZONE,
    delivery_window_end TIMESTAMP WITH TIME ZONE,
    activated_at TIMESTAMP WITH TIME ZONE,
    region_id UUID NOT NULL REFERENCES regions(id) ON DELETE CASCADE,
    currency_code VARCHAR(3) NOT NULL,
    tax_zone_id UUID REFERENCES tax_zones(id) ON DELETE SET NULL,
    pricing_mode VARCHAR(20) NOT NULL DEFAULT 'tax_exclusive' CHECK (pricing_mode IN ('tax_inclusive', 'tax_exclusive')),
    subtotal_amount INTEGER NOT NULL CHECK (subtotal_amount >= 0),
    delivery_fee_amount INTEGER NOT NULL DEFAULT 0 CHECK (delivery_fee_amount >= 0),
    tax_amount INTEGER NOT NULL CHECK (tax_amount >= 0),
    total_amount INTEGER NOT NULL CHECK (total_amount >= 0),
    delivery_address_street TEXT NOT NULL,
    delivery_address_city VARCHAR(100) NOT NULL,
    delivery_address_state VARCHAR(100),
    delivery_address_postal_code VARCHAR(20) NOT NULL,
    delivery_address_country_code VARCHAR(2) NOT NULL,
    delivery_address_latitude DECIMAL(10, 8),
    delivery_address_longitude DECIMAL(11, 8),
    notes TEXT,
    cancellation_reason TEXT,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Order items table
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id UUID NOT NULL REFERENCES menu_items(id) ON DELETE SET NULL,
    item_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_amount INTEGER NOT NULL CHECK (unit_price_amount >= 0),
    line_subtotal_amount INTEGER NOT NULL CHECK (line_subtotal_amount >= 0),
    line_tax_amount INTEGER NOT NULL DEFAULT 0 CHECK (line_tax_amount >= 0),
    line_total_amount INTEGER NOT NULL CHECK (line_total_amount >= 0),
    currency_code VARCHAR(3) NOT NULL,
    tax_category_id UUID REFERENCES tax_categories(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Order tax lines table (tax snapshot)
CREATE TABLE order_tax_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    tax_category_id UUID REFERENCES tax_categories(id) ON DELETE SET NULL,
    tax_rate_name VARCHAR(100) NOT NULL,
    tax_rate_bps INTEGER NOT NULL,
    is_inclusive BOOLEAN NOT NULL,
    taxable_amount INTEGER NOT NULL CHECK (taxable_amount >= 0),
    tax_amount INTEGER NOT NULL CHECK (tax_amount >= 0),
    currency_code VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Payments table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount_amount INTEGER NOT NULL CHECK (amount_amount >= 0),
    currency_code VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'processing', 'succeeded', 'failed', 'refunded')),
    payment_method VARCHAR(50),
    provider_payment_id VARCHAR(255),
    provider_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Deliveries table
CREATE TABLE deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    courier_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending_schedule', 'unassigned', 'assigned', 'picked_up', 'delivered', 'failed', 'cancelled')),
    pickup_address_street TEXT NOT NULL,
    pickup_address_city VARCHAR(100) NOT NULL,
    pickup_address_state VARCHAR(100),
    pickup_address_postal_code VARCHAR(20) NOT NULL,
    pickup_address_country_code VARCHAR(2) NOT NULL,
    pickup_address_latitude DECIMAL(10, 8),
    pickup_address_longitude DECIMAL(11, 8),
    delivery_address_street TEXT NOT NULL,
    delivery_address_city VARCHAR(100) NOT NULL,
    delivery_address_state VARCHAR(100),
    delivery_address_postal_code VARCHAR(20) NOT NULL,
    delivery_address_country_code VARCHAR(2) NOT NULL,
    delivery_address_latitude DECIMAL(10, 8),
    delivery_address_longitude DECIMAL(11, 8),
    estimated_pickup_at TIMESTAMP WITH TIME ZONE,
    estimated_delivery_at TIMESTAMP WITH TIME ZONE,
    assigned_at TIMESTAMP WITH TIME ZONE,
    picked_up_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    cancellation_reason TEXT,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Couriers table
CREATE TABLE couriers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone VARCHAR(20),
    vehicle_type VARCHAR(50),
    is_available BOOLEAN NOT NULL DEFAULT true,
    current_latitude DECIMAL(10, 8),
    current_longitude DECIMAL(11, 8),
    current_location_updated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Outbox events table
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id UUID NOT NULL,
    payload JSONB NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT false,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Jobs table
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_type VARCHAR(50) NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id UUID NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'succeeded', 'failed', 'dead_letter')),
    priority INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 5,
    locked_at TIMESTAMP WITH TIME ZONE,
    locked_by VARCHAR(100),
    last_error TEXT,
    processed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Idempotency keys table
CREATE TABLE idempotency_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) NOT NULL,
    scope VARCHAR(100) NOT NULL,
    request_hash VARCHAR(64),
    response_code INTEGER,
    response_body JSONB,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    UNIQUE (key, scope)
);

-- Audit logs table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_id UUID NOT NULL,
    actor_role VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID,
    metadata JSONB,
    ip_address INET,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Indexes

-- Users indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Refresh tokens indexes
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- User addresses indexes
CREATE INDEX idx_user_addresses_user_id ON user_addresses(user_id);
CREATE INDEX idx_user_addresses_deleted_at ON user_addresses(deleted_at) WHERE deleted_at IS NOT NULL;

-- Regions indexes
CREATE INDEX idx_regions_code ON regions(code);
CREATE INDEX idx_regions_country_code ON regions(country_code);
CREATE INDEX idx_regions_is_active ON regions(is_active);
CREATE INDEX idx_regions_deleted_at ON regions(deleted_at) WHERE deleted_at IS NOT NULL;

-- Tax zones indexes
CREATE INDEX idx_tax_zones_country_code ON tax_zones(country_code);
CREATE INDEX idx_tax_zones_state ON tax_zones(state_or_province);
CREATE INDEX idx_tax_zones_city ON tax_zones(city);
CREATE INDEX idx_tax_zones_postal_pattern ON tax_zones(postal_code_pattern);
CREATE INDEX idx_tax_zones_is_active ON tax_zones(is_active);
CREATE INDEX idx_tax_zones_deleted_at ON tax_zones(deleted_at) WHERE deleted_at IS NOT NULL;

-- Tax rates indexes
CREATE INDEX idx_tax_rates_tax_zone_id ON tax_rates(tax_zone_id);
CREATE INDEX idx_tax_rates_tax_category_id ON tax_rates(tax_category_id);
CREATE INDEX idx_tax_rates_effective_dates ON tax_rates(effective_from, effective_to);
CREATE INDEX idx_tax_rates_is_active ON tax_rates(is_active);
CREATE INDEX idx_tax_rates_deleted_at ON tax_rates(deleted_at) WHERE deleted_at IS NOT NULL;

-- Restaurants indexes
CREATE INDEX idx_restaurants_owner_id ON restaurants(owner_id);
CREATE INDEX idx_restaurants_region_id ON restaurants(region_id);
CREATE INDEX idx_restaurants_is_active ON restaurants(is_active);
CREATE INDEX idx_restaurants_is_approved ON restaurants(is_approved);
CREATE INDEX idx_restaurants_location ON restaurants(latitude, longitude);
CREATE INDEX idx_restaurants_deleted_at ON restaurants(deleted_at) WHERE deleted_at IS NOT NULL;

-- Menu items indexes
CREATE INDEX idx_menu_items_restaurant_id ON menu_items(restaurant_id);
CREATE INDEX idx_menu_items_tax_category_id ON menu_items(tax_category_id);
CREATE INDEX idx_menu_items_is_available ON menu_items(is_available);
CREATE INDEX idx_menu_items_is_active ON menu_items(is_active);
CREATE INDEX idx_menu_items_deleted_at ON menu_items(deleted_at) WHERE deleted_at IS NOT NULL;

-- Orders indexes
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_restaurant_id ON orders(restaurant_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_region_id ON orders(region_id);
CREATE INDEX idx_orders_tax_zone_id ON orders(tax_zone_id);
CREATE INDEX idx_orders_scheduled_for ON orders(scheduled_for) WHERE scheduled_for IS NOT NULL;
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_deleted_at ON orders(deleted_at) WHERE deleted_at IS NOT NULL;

-- Order items indexes
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_menu_item_id ON order_items(menu_item_id);

-- Order tax lines indexes
CREATE INDEX idx_order_tax_lines_order_id ON order_tax_lines(order_id);

-- Payments indexes
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_customer_id ON payments(customer_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at DESC);

-- Deliveries indexes
CREATE INDEX idx_deliveries_order_id ON deliveries(order_id);
CREATE INDEX idx_deliveries_courier_id ON deliveries(courier_id);
CREATE INDEX idx_deliveries_status ON deliveries(status);
CREATE INDEX idx_deliveries_created_at ON deliveries(created_at DESC);

-- Couriers indexes
CREATE INDEX idx_couriers_user_id ON couriers(user_id);
CREATE INDEX idx_couriers_is_available ON couriers(is_available);
CREATE INDEX idx_couriers_location ON couriers(current_latitude, current_longitude);

-- Outbox events indexes
CREATE INDEX idx_outbox_events_processed ON outbox_events(processed, created_at);
CREATE INDEX idx_outbox_events_aggregate ON outbox_events(aggregate_type, aggregate_id);

-- Jobs indexes
CREATE INDEX idx_jobs_status ON jobs(status, next_attempt_at);
CREATE INDEX idx_jobs_aggregate ON jobs(aggregate_type, aggregate_id);
CREATE INDEX idx_jobs_job_type ON jobs(job_type);
CREATE INDEX idx_jobs_priority ON jobs(priority DESC, next_attempt_at ASC);

-- Idempotency keys indexes
CREATE INDEX idx_idempotency_keys_key_scope ON idempotency_keys(key, scope);
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);

-- Audit logs indexes
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Updated_at triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_addresses_updated_at BEFORE UPDATE ON user_addresses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_regions_updated_at BEFORE UPDATE ON regions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_region_configs_updated_at BEFORE UPDATE ON region_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tax_categories_updated_at BEFORE UPDATE ON tax_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tax_zones_updated_at BEFORE UPDATE ON tax_zones
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tax_rates_updated_at BEFORE UPDATE ON tax_rates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_restaurants_updated_at BEFORE UPDATE ON restaurants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_restaurant_hours_updated_at BEFORE UPDATE ON restaurant_hours
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_menu_items_updated_at BEFORE UPDATE ON menu_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_order_items_updated_at BEFORE UPDATE ON order_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_deliveries_updated_at BEFORE UPDATE ON deliveries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_couriers_updated_at BEFORE UPDATE ON couriers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_jobs_updated_at BEFORE UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
