BEGIN;
CREATE TABLE IF NOT EXISTS public.users_balance (
    id bigint GENERATED ALWAYS AS IDENTITY NOT NULL,
    "user_id" bigint NOT NULL,
    "current" float NOT NULL DEFAULT 0,
    "withdrawn" float NOT NULL DEFAULT 0,
    "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_balance_pk PRIMARY KEY (id),
    CONSTRAINT users_balance_unique UNIQUE (user_id),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id)
);
COMMIT;