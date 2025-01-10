BEGIN;
CREATE TABLE IF NOT EXISTS public.withdrawals (
    id bigint GENERATED ALWAYS AS IDENTITY NOT NULL,
    "order_id" bigint NOT NULL,
    "sum" float NOT NULL DEFAULT 0,
    "processed_at" timestamp NULL,
    "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT withdrawals_pk PRIMARY KEY (id),
    CONSTRAINT fk_order FOREIGN KEY(order_id) REFERENCES orders(id)
);
COMMIT;