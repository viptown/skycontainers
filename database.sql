CREATE TABLE "container_types"(
    "id" BIGINT NOT NULL,
    "code" VARCHAR(10) NOT NULL,
    "length_ft" SMALLINT NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "soft_order" INTEGER NOT NULL,
    "created_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL,
        "updated_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL
);
ALTER TABLE
    "container_types" ADD CONSTRAINT "container_types_code_unique" UNIQUE("code");
ALTER TABLE
    "container_types" ADD PRIMARY KEY("id");
CREATE TABLE "containers"(
    "id" BIGINT NOT NULL,
    "containers_type_id" BIGINT NOT NULL,
    "container_no" VARCHAR(255) NOT NULL,
    "container_status" VARCHAR(255) CHECK
        (
            "container_status" IN('empty', 'full', 'store')
        ) NOT NULL DEFAULT 'empty',
        "supplier_id" BIGINT NOT NULL,
        "booking_no" VARCHAR(30) NOT NULL,
        "car_no" VARCHAR(30) NOT NULL,
        "memo" TEXT NOT NULL,
        "user_id" BIGINT NOT NULL,
        "inbound_date" DATE NOT NULL,
        "processing_date" DATE NOT NULL,
        "outbound_date" DATE NOT NULL,
        "processing_cancelled_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL,
        "processing_cancelled_by" BIGINT NOT NULL,
        "created_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL,
        "updated_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL
);
CREATE INDEX "containers_container_no_index" ON
    "containers"("container_no");
ALTER TABLE
    "containers" ADD PRIMARY KEY("id");
COMMENT
ON COLUMN
    "containers"."container_status" IS 'Empty,Full,보관';
COMMENT
ON COLUMN
    "containers"."inbound_date" IS '입고일';
COMMENT
ON COLUMN
    "containers"."processing_date" IS '작업일';
COMMENT
ON COLUMN
    "containers"."outbound_date" IS '출고일';
CREATE TABLE "suppliers"(
    "id" BIGINT NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "tel" VARCHAR(255) NOT NULL,
    "email" VARCHAR(255) NOT NULL,
    "is_active" BOOLEAN NOT NULL,
    "created_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL,
        "updated_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL
);
ALTER TABLE
    "suppliers" ADD PRIMARY KEY("id");
CREATE TABLE "users"(
    "id" BIGINT NOT NULL,
    "supplier_id" BIGINT,
    "uid" VARCHAR(50) NOT NULL,
    "password_hash" TEXT NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "email" VARCHAR(255) NOT NULL,
    "duty" VARCHAR(255) CHECK
        (
            "duty" IN(
                'Senior',
                'Assistant',
                'Manager',
                'Senior Manager',
                'director',
                'other'
            )
        ) NOT NULL,
        "phone" VARCHAR(30) NOT NULL,
        "role" VARCHAR(40) NOT NULL,
        "status" VARCHAR(255)
    CHECK
        (
            "status" IN('active', 'suspened', 'deleted')
        ) NOT NULL,
        "last_login_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL,
        "created_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL,
        "updated_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL
);
CREATE INDEX "users_supplier_id_index" ON
    "users"("supplier_id");
ALTER TABLE
    "users" ADD PRIMARY KEY("id");
COMMENT
ON COLUMN
    "users"."duty" IS '직원,주임,대리,과장,차장,이사,기타';
CREATE TABLE "bl_markings"(
    "id" BIGINT NOT NULL,
    "container_id" BIGINT NOT NULL,
    "user_id" BIGINT NOT NULL,
    "hbl_no" VARCHAR(255) NOT NULL,
    "marks" VARCHAR(255) NOT NULL,
    "is_active" BOOLEAN NOT NULL,
    "created_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL,
        "updated_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL,
        "bl_position_id" BIGINT
);
ALTER TABLE
    "bl_markings" ADD PRIMARY KEY("id");
CREATE INDEX "bl_markings_hbl_no_index" ON
    "bl_markings"("hbl_no");
CREATE TABLE "carnumbers"(
    "id" BIGINT NOT NULL,
    "log_date" VARCHAR(15) NOT NULL,
    "car_no" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL
);
ALTER TABLE
    "carnumbers" ADD PRIMARY KEY("id");
CREATE TABLE "reports"(
    "id" BIGINT NOT NULL,
    "user_id" BIGINT NOT NULL,
    "subject" VARCHAR(255) NOT NULL,
    "contents" TEXT NOT NULL,
    "types" VARCHAR(255) CHECK
        ("types" IN('1', '2', '3', '4', '5', '6')) NOT NULL DEFAULT '1',
        "period_start" DATE NOT NULL,
        "period_end" DATE NOT NULL,
        "is_active" BOOLEAN NOT NULL,
        "created_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL,
        "updated_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL
);
ALTER TABLE
    "reports" ADD PRIMARY KEY("id");
COMMENT
ON COLUMN
    "reports"."types" IS '반차,연차,경조,병가,무급,기타';
CREATE TABLE "policy_permissions"(
    "role" VARCHAR(40) NOT NULL,
    "resource" VARCHAR(100) NOT NULL,
    "action" VARCHAR(20) NOT NULL,
    "allowed" BOOLEAN NOT NULL,
    "created_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL,
        "updated_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL
);
ALTER TABLE
    "policy_permissions" ADD PRIMARY KEY("role", "resource", "action");
CREATE TABLE "bl_positions"(
    "id" BIGINT NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "is_active" BOOLEAN NOT NULL DEFAULT true,
    "created_at" TIMESTAMP(0) WITH
        TIME zone NOT NULL,
        "updated_at" TIMESTAMP(0)
    WITH
        TIME zone NOT NULL,
        "user_id" BIGINT NOT NULL
);
ALTER TABLE
    "bl_positions" ADD PRIMARY KEY("id");
ALTER TABLE
    "users" ADD CONSTRAINT "users_supplier_id_foreign" FOREIGN KEY("supplier_id") REFERENCES "suppliers"("id");
ALTER TABLE
    "containers" ADD CONSTRAINT "containers_containers_type_id_foreign" FOREIGN KEY("containers_type_id") REFERENCES "container_types"("id");
ALTER TABLE
    "containers" ADD CONSTRAINT "containers_processing_cancelled_by_foreign" FOREIGN KEY("processing_cancelled_by") REFERENCES "users"("id");
ALTER TABLE
    "containers" ADD CONSTRAINT "containers_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "users"("id");
ALTER TABLE
    "bl_positions" ADD CONSTRAINT "bl_positions_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "users"("id");
ALTER TABLE
    "reports" ADD CONSTRAINT "reports_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "users"("id");
ALTER TABLE
    "bl_markings" ADD CONSTRAINT "bl_markings_container_id_foreign" FOREIGN KEY("container_id") REFERENCES "containers"("id");
ALTER TABLE
    "bl_markings" ADD CONSTRAINT "bl_markings_bl_position_id_foreign" FOREIGN KEY("bl_position_id") REFERENCES "bl_positions"("id");
ALTER TABLE
    "bl_markings" ADD CONSTRAINT "bl_markings_user_id_foreign" FOREIGN KEY("user_id") REFERENCES "users"("id");
ALTER TABLE
    "containers" ADD CONSTRAINT "containers_supplier_id_foreign" FOREIGN KEY("supplier_id") REFERENCES "suppliers"("id");
