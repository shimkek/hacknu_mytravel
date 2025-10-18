-- Simple PostgreSQL database initialization
-- Structure based on project requirements

-- Create enum types
CREATE TYPE accommodation_type AS ENUM (
    'hotel', 'hostel', 'apartment', 'guest_house', 'resort', 'camping', 'villa', 'other'
);

CREATE TYPE verification_status AS ENUM (
    'new',           -- новый
    'verified',      -- проверен  
    'in_development' -- в разработке
);

CREATE TYPE source_website AS ENUM (
    '2gis', 'google_maps', 'instagram', 'olx', 'yandex', 'manual'
);

-- Main accommodations table based on project requirements
CREATE TABLE accommodations (
    id SERIAL PRIMARY KEY,
    
    -- Название объекта
    name VARCHAR(500) NOT NULL,
    
    -- Координаты (GPS)
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    
    -- Адрес
    address TEXT,
    
    -- Тип размещения (категория)
    accommodation_type accommodation_type,
    
    -- Контактные данные (телефон, email, соцсети)
    phone VARCHAR(50),
    email VARCHAR(100),
    social_media_links JSONB, -- соцсети как JSON объект
    
    -- Сайт/страница в соцсети
    website_url TEXT,
    social_media_page TEXT, -- основная страница в соцсетях
    
    -- Описание услуг
    service_description TEXT,
    
    -- Количество номеров/мест
    room_count INTEGER,
    capacity INTEGER, -- общее количество мест
    
    -- Ценовой диапазон
    price_range_min DECIMAL(10, 2),
    price_range_max DECIMAL(10, 2),
    price_currency VARCHAR(3) DEFAULT 'KZT',
    
    -- Фотографии (ссылки или мини-галерея)
    photos JSONB, -- массив ссылок на фотографии
    
    -- Отзывы и рейтинги (если есть)
    rating DECIMAL(3, 2), -- рейтинг от 0.00 до 5.00
    review_count INTEGER DEFAULT 0,
    reviews JSONB, -- массив отзывов
    
    -- Инфраструктура (Wi-Fi, паркинг, кухня и т.д.)
    amenities JSONB, -- объект с удобствами
    
    -- Статус проверки (новый, проверен, в разработке)
    verification_status verification_status DEFAULT 'new',
    
    -- Дата последнего обновления
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Технические поля для системы
    source_website source_website NOT NULL,
    source_url TEXT, -- URL источника данных
    external_id VARCHAR(100), -- ID объекта на источнике
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    
    CONSTRAINT unique_source_external_id UNIQUE (source_website, external_id)
);

-- Parsing logs table
CREATE TABLE parsing_logs (
    id SERIAL PRIMARY KEY,
    source_website source_website NOT NULL,
    status VARCHAR(20) NOT NULL,
    records_processed INTEGER DEFAULT 0,
    records_inserted INTEGER DEFAULT 0,
    records_updated INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER
);

-- Create indexes for better performance
CREATE INDEX idx_accommodations_type ON accommodations (accommodation_type);
CREATE INDEX idx_accommodations_source ON accommodations (source_website);
CREATE INDEX idx_accommodations_status ON accommodations (verification_status);
CREATE INDEX idx_accommodations_created_at ON accommodations (created_at);
CREATE INDEX idx_accommodations_last_updated ON accommodations (last_updated);
CREATE INDEX idx_accommodations_location ON accommodations (latitude, longitude);

-- Auto-update last_updated timestamp function
CREATE OR REPLACE FUNCTION update_last_updated_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_updated = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for auto-updating last_updated field (matching project requirements)
CREATE TRIGGER update_accommodations_last_updated 
    BEFORE UPDATE ON accommodations 
    FOR EACH ROW 
    EXECUTE FUNCTION update_last_updated_column();

-- Sample data with correct structure
INSERT INTO accommodations (
    name, 
    latitude, longitude, 
    address, 
    accommodation_type,
    phone, 
    email,
    website_url,
    service_description,
    room_count, 
    capacity, 
    price_range_min, 
    price_range_max,
    rating, 
    review_count,
    amenities,
    verification_status,
    source_website,
    external_id
) VALUES 
(
    'Отель Алматы Центр', 
    43.2220, 76.8512, 
    'пр. Достык 123, Алматы, Казахстан',
    'hotel',
    '+7 727 123 4567',
    'info@almaty-center.kz',
    'https://almaty-center.kz',
    'Комфортабельный отель в центре Алматы с полным спектром услуг',
    50, 100, 
    15000.00, 45000.00,
    4.2, 127,
    '{"wifi": true, "parking": true, "breakfast": true, "gym": true, "pool": false}',
    'verified',
    'manual',
    'sample_001'
),
(
    'Хостел Астана Бюджет',
    51.1694, 71.4491,
    'пр. Мангилик Ел 456, Астана, Казахстан', 
    'hostel',
    '+7 717 987 6543',
    'info@astana-budget.kz',
    'https://astana-budget.kz',
    'Бюджетное размещение для путешественников в центре столицы',
    10, 40, 
    3000.00, 8000.00,
    4.0, 89,
    '{"wifi": true, "parking": false, "kitchen": true, "laundry": true, "24h_reception": true}',
    'verified',
    'manual',
    'sample_002'
);