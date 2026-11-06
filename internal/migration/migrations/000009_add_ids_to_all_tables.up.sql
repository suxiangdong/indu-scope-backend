-- Give every business table an independent, auto-incrementing primary key.
-- Existing natural/relationship uniqueness is preserved with unique keys.

ALTER TABLE personal_verifications
    ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (id),
    ADD UNIQUE KEY uk_personal_verifications_user (user_id);

ALTER TABLE company_verifications
    ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (id),
    ADD UNIQUE KEY uk_company_verifications_user (user_id);

ALTER TABLE company_profiles
    ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (id),
    ADD UNIQUE KEY uk_company_profiles_user (user_id);

ALTER TABLE category_attributes
    ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (id),
    ADD UNIQUE KEY uk_category_attributes_category_attribute (category_id, attribute_id);

ALTER TABLE product_attribute_value_options
    ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (id),
    ADD UNIQUE KEY uk_product_attribute_value_options_value_option (value_id, option_id);

ALTER TABLE attribute_group_attributes
    ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (id),
    ADD UNIQUE KEY uk_attribute_group_attributes_group_attribute (group_id, attribute_id);

ALTER TABLE category_attribute_group_bindings
    ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (id),
    ADD UNIQUE KEY uk_category_attribute_group_bindings_category_group (category_id, group_id);
