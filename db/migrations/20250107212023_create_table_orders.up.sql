BEGIN;
CREATE TABLE IF NOT EXISTS public.orders (
    id bigint GENERATED ALWAYS AS IDENTITY NOT NULL,
    "user_id" bigint NOT NULL,
    "number" varchar(255) NOT NULL,
    "status" smallint NOT NULL,
    "accrual" float NULL,
    "uploaded_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT orders_pk PRIMARY KEY (id),
    CONSTRAINT orders_unique UNIQUE (number),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id)
);
COMMIT;