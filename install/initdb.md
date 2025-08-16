```sql
-- SQL used to init the DB

CREATE TABLE vendor (
    id integer primary key autoincrement,
    name text not null,
    unique (name)
);

CREATE TABLE user (
    id integer primary key autoincrement,
    name text not null, 
    isadmin int not null default 0,
    pwdhash text not null,
    unique (name),
    check (case id when 1 then isadmin else 1 end = 1)
);

CREATE TABLE ds (
    id integer primary key autoincrement,
    name text not null, 
    location text not null, 
    vendor_id text not null,
    unique (name),
    foreign key(vendor_id) references vendor(id)
);

CREATE TABLE user_ds (
    id integer primary key autoincrement, 
    user_id int not null, 
    ds_id int not null, 
    unique(user_id, ds_id),
    foreign key(user_id) references user(id),
    foreign key(ds_id) references ds(id)
);

-- init DB data
--
-- add vendor list
insert into vendor (name) values ('sqlite3'), ('clickhouse'), ('mssql'), ('mysql'), ('postgresql');

-- add admin user. (pwdhash is set to an empty string hash. Which makes it impossible to use for login.)
insert into user (name, isadmin, pwdhash) values ('admin', 1, '$2a$10$uu2BBL5jm9/GhUvLmuxcVO4pKLTIzf8jOl4HV9bTUu2Ss203eDNJK');

-- add internal data source
insert into ds (name, vendor_id, location) values ('db-portal', 1, 'n/a');

-- give admin access to internal data source
insert into user_ds (user_id, ds_id) values (1, 1);

```