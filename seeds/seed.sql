-- Seed Data for Hopper Food Delivery API Backend
-- This script creates initial data for development/testing

-- Note: Passwords are hashed with bcrypt. 
-- Default password for all test users: SecurePass123!

-- Insert Admin User
INSERT INTO users (id, email, password_hash, role, first_name, last_name)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin@hopper.example.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYlqJq9Q5yi', -- bcrypt hash for "SecurePass123!"
    'admin',
    'System',
    'Administrator'
);

-- Insert Test Customer
INSERT INTO users (id, email, password_hash, role, first_name, last_name)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'customer@example.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYlqJq9Q5yi',
    'customer',
    'John',
    'Doe'
);

-- Insert Test Restaurant Owner
INSERT INTO users (id, email, password_hash, role, first_name, last_name)
VALUES (
    '00000000-0000-0000-0000-000000000003',
    'owner@example.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYlqJq9Q5yi',
    'restaurant_owner',
    'Jane',
    'Smith'
);

-- Insert Test Courier
INSERT INTO users (id, email, password_hash, role, first_name, last_name)
VALUES (
    '00000000-0000-0000-0000-000000000004',
    'courier@example.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYlqJq9Q5yi',
    'courier',
    'Bob',
    'Johnson'
);

-- Insert Courier Profile
INSERT INTO couriers (id, user_id, phone, vehicle_type, is_available)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000004',
    '+14155551234',
    'bicycle',
    true
);

-- Insert Region (US - California)
INSERT INTO regions (id, code, name, country_code, timezone, currency_code)
VALUES (
    '00000000-0000-0000-0000-000000000010',
    'US-CA',
    'California',
    'US',
    'America/Los_Angeles',
    'USD'
);

-- Insert Region Config for California
INSERT INTO region_configs (id, region_id, platform_fee_basis_points, default_delivery_window_minutes, order_activation_lead_minutes, allow_scheduled_orders, delivery_fee_taxable_default)
VALUES (
    '00000000-0000-0000-0000-000000000011',
    '00000000-0000-0000-0000-000000000010',
    1500,
    30,
    15,
    true,
    true
);

-- Insert Tax Categories
INSERT INTO tax_categories (id, code, name, description) VALUES
('00000000-0000-0000-0000-000000000020', 'food', 'Food and Beverages', 'Standard food items'),
('00000000-0000-0000-0000-000000000021', 'alcohol', 'Alcoholic Beverages', 'Alcoholic beverages'),
('00000000-0000-0000-0000-000000000022', 'delivery', 'Delivery Fee', 'Delivery service fees');

-- Insert Tax Zone (California)
INSERT INTO tax_zones (id, code, name, country_code, state_or_province)
VALUES (
    '00000000-0000-0000-0000-000000000030',
    'US-CA',
    'California',
    'US',
    'CA'
);

-- Insert Tax Rates (California Sales Tax - 8.25%)
INSERT INTO tax_rates (id, tax_zone_id, tax_category_id, name, rate_bps, is_inclusive, applies_to_delivery_fee)
VALUES
('00000000-0000-0000-0000-000000000040', '00000000-0000-0000-0000-000000000030', '00000000-0000-0000-0000-000000000020', 'California Sales Tax', 825, false, true),
('00000000-0000-0000-0000-000000000041', '00000000-0000-0000-0000-000000000030', '00000000-0000-0000-0000-000000000021', 'California Alcohol Tax', 825, false, false),
('00000000-0000-0000-0000-000000000042', '00000000-0000-0000-0000-000000000030', '00000000-0000-0000-0000-000000000022', 'California Delivery Tax', 825, false, true);

-- Insert Restaurant
INSERT INTO restaurants (id, owner_id, name, description, cuisine_type, street_address, city, state_or_province, postal_code, country_code, latitude, longitude, phone, email, region_id, currency_code, timezone, is_active, is_approved)
VALUES (
    '00000000-0000-0000-0000-000000000100',
    '00000000-0000-0000-0000-000000000003',
    'Tasty Burgers',
    'Best burgers in town',
    'American',
    '123 Main St',
    'San Francisco',
    'CA',
    '94102',
    'US',
    37.7749,
    -122.4194,
    '+14155555555',
    'info@tastyburgers.com',
    '00000000-0000-0000-0000-000000000010',
    'USD',
    'America/Los_Angeles',
    true,
    true
);

-- Insert Restaurant Hours (Mon-Sun 10am-10pm)
INSERT INTO restaurant_hours (id, restaurant_id, day_of_week, open_time, close_time, is_closed) VALUES
('00000000-0000-0000-0000-000000000110', '00000000-0000-0000-0000-000000000100', 0, '10:00', '22:00', false),
('00000000-0000-0000-0000-000000000111', '00000000-0000-0000-0000-000000000100', 1, '10:00', '22:00', false),
('00000000-0000-0000-0000-000000000112', '00000000-0000-0000-0000-000000000100', 2, '10:00', '22:00', false),
('00000000-0000-0000-0000-000000000113', '00000000-0000-0000-0000-000000000100', 3, '10:00', '22:00', false),
('00000000-0000-0000-0000-000000000114', '00000000-0000-0000-0000-000000000100', 4, '10:00', '22:00', false),
('00000000-0000-0000-0000-000000000115', '00000000-0000-0000-0000-000000000100', 5, '10:00', '22:00', false),
('00000000-0000-0000-0000-000000000116', '00000000-0000-0000-0000-000000000100', 6, '10:00', '22:00', false);

-- Insert Menu Items
INSERT INTO menu_items (id, restaurant_id, name, description, price_amount, currency_code, tax_category_id, is_available, is_active) VALUES
('00000000-0000-0000-0000-000000000200', '00000000-0000-0000-0000-000000000100', 'Classic Burger', 'Beef patty with lettuce, tomato, onion', 1200, 'USD', '00000000-0000-0000-0000-000000000020', true, true),
('00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000100', 'Cheese Burger', 'Beef patty with cheddar cheese', 1400, 'USD', '00000000-0000-0000-0000-000000000020', true, true),
('00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000100', 'Bacon Burger', 'Beef patty with crispy bacon', 1600, 'USD', '00000000-0000-0000-0000-000000000020', true, true),
('00000000-0000-0000-0000-000000000203', '00000000-0000-0000-0000-000000000100', 'French Fries', 'Crispy golden fries', 400, 'USD', '00000000-0000-0000-0000-000000000020', true, true),
('00000000-0000-0000-0000-000000000204', '00000000-0000-0000-0000-000000000100', 'Soda', 'Choice of cola, sprite, or lemonade', 300, 'USD', '00000000-0000-0000-0000-000000000020', true, true);

-- Insert Customer Address
INSERT INTO user_addresses (id, user_id, street_address, city, state_or_province, postal_code, country_code, latitude, longitude, is_default)
VALUES (
    '00000000-0000-0000-0000-000000000300',
    '00000000-0000-0000-0000-000000000002',
    '456 Oak Street',
    'San Francisco',
    'CA',
    '94103',
    'US',
    37.7649,
    -122.4294,
    true
);
