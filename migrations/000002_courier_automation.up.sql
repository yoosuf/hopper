CREATE TABLE courier_location_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    courier_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX idx_courier_location_events_courier_time
    ON courier_location_events(courier_user_id, created_at DESC);
