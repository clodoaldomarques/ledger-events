CREATE TABLE IF NOT EXISTS event(
    event_id varchar(50) primary key,
    correlation_id varchar(50) not null,
    org_id varchar(60) not null,
    event_type_id varchar(100),
    transaction_id bigint,
    authorization_id bigint,
    program_id bigint not null,
    customer_id bigint,
    account_id bigint not null,
    days_due bigint,
    description varchar(200) not null,
    transaction_type bigint,
    process_id bigint,
    producer varchar(100) not null,
    processing_code varchar(100),
    processing_description varchar(200),
    clearing_date timestamp,
    chunk_part bigint,
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE IF NOT EXISTS entry(
    event_id varchar(50) not null,
    event_type_id bigint not null,
    amount decimal(20, 15) not null,
    debit_account varchar(60),
    credit_account varchar(60),
    cost_center_debit_org varchar(60),
    cost_center_debit_cost varchar(60),
    cost_center_credit_org varchar(60),
    cost_center_credit_cost varchar(60),
    company_code varchar(50),
    company_type varchar(50),
    description varchar(200) not null,
    created_at timestamp not null,
    primary key(event_id, event_type_id)
);