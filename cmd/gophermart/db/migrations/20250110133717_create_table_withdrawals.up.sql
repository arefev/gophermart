BEGIN;
CREATE TABLE IF NOT EXISTS public.withdrawals (
    id bigint GENERATED ALWAYS AS IDENTITY NOT NULL,
    "user_id" bigint NOT NULL,
    "number" varchar(255) NOT NULL,
    "sum" float NOT NULL DEFAULT 0,
    "processed_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT withdrawals_pk PRIMARY KEY (id),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id)
);
COMMIT;