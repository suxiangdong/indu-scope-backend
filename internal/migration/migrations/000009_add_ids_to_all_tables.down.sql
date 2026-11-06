ALTER TABLE category_attribute_group_bindings
    DROP PRIMARY KEY,
    DROP KEY uk_category_attribute_group_bindings_category_group,
    DROP COLUMN id,
    ADD PRIMARY KEY (category_id, group_id);

ALTER TABLE attribute_group_attributes
    DROP PRIMARY KEY,
    DROP KEY uk_attribute_group_attributes_group_attribute,
    DROP COLUMN id,
    ADD PRIMARY KEY (group_id, attribute_id);

ALTER TABLE product_attribute_value_options
    DROP PRIMARY KEY,
    DROP KEY uk_product_attribute_value_options_value_option,
    DROP COLUMN id,
    ADD PRIMARY KEY (value_id, option_id);

ALTER TABLE category_attributes
    DROP PRIMARY KEY,
    DROP KEY uk_category_attributes_category_attribute,
    DROP COLUMN id,
    ADD PRIMARY KEY (category_id, attribute_id);

ALTER TABLE company_profiles
    DROP PRIMARY KEY,
    DROP KEY uk_company_profiles_user,
    DROP COLUMN id,
    ADD PRIMARY KEY (user_id);

ALTER TABLE company_verifications
    DROP PRIMARY KEY,
    DROP KEY uk_company_verifications_user,
    DROP COLUMN id,
    ADD PRIMARY KEY (user_id);

ALTER TABLE personal_verifications
    DROP PRIMARY KEY,
    DROP KEY uk_personal_verifications_user,
    DROP COLUMN id,
    ADD PRIMARY KEY (user_id);
