CREATE TABLE public.users (
    id serial PRIMARY KEY,
    username varchar(255),
    email varchar(255) UNIQUE not null,
    password varchar(512) not null
);

CREATE TABLE public.keys (
    user_id integer UNIQUE references public.users(id) primary key ,
    api_key varchar(512),
    secret_key varchar(512)
);

CREATE TABLE public.strategy_info (
    id serial PRIMARY KEY,
    name varchar(255),
    description varchar(1000)
);

INSERT INTO public.strategy_info(name, description) VALUES ('Moving Average Crossover','A simple moving average crossover strategy. Buy if short term MA is higher than long term MA. Sell if short term MA below long term MA.');
INSERT INTO public.strategy_info(id, name, description) VALUES (0, 'DEFAULT_TEMPLATE','');

CREATE TABLE public.strategy_fields(
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER not null references public.strategy_info(id),
    name VARCHAR(150) NOT NULL,
    display_name VARCHAR(150) NOT NULL,
    description VARCHAR(512),
    min INTEGER,
    max INTEGER,
    default_value VARCHAR(512) NOT NULL ,
    type VARCHAR(50) NOT NULL ,
    ui_type VARCHAR(50) NOT NULL,
    dataset varchar
);

INSERT INTO public.strategy_fields(strategy_id, name, display_name, description, min, max, default_value, type, ui_type, dataset)
    VALUES (0,'pair','Pair','Some of Spot trading pairs may be unavailable on Futures and vise versa.', NULL, NULL, 'BTCUSDT','string','modal-select','getAllPairs'),
           (0,'bid','Bid Size','Minimal bid on Binance is 11$', 11, NULL, '11','number','input',NULL),
           (0,'timeFrame','Time Frame Duration','Duration of time frame equals to one bar on barchart.', NULL, NULL, '1','number','select','1 minute:1!!3 minutes:3!!5 minutes:5!!15 minutes:15!!30 minutes:30!!1 hour:60'),
           (0, 'isFutures', 'Trading futures',' Leverage of asset will be set to binance default. Go to binance to change it.', NULL, NULL, 'false','bool','checkbox', NULL),
           (1,'longTerm','Long Term Period','Number of periods for long term moving average calculation.',2,NULL,25,'number','input',NULL),
           (2,'shortTerm','Short Term Period','Number of periods for short term moving average calculation.',1,NULL,7,'number','input', NULL);

CREATE TABLE strategy_instances (
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER NOT NULL references public.strategy_info(id),
    user_id INTEGER NOT NULL references public.users(id),
    pair VARCHAR(20) NOT NULL,
    bid numeric NOT NULL,
    time_frame integer NOT NULL,
    status varchar(20) NOT NULL
);

CREATE TABLE public.trades (
    id SERIAL primary key,
    pair varchar(20) not null,
    strategy_id integer references public.strategy_info(id),
    instance_id integer references public.strategy_instances(id) ON DELETE CASCADE,
    user_id integer references public.users(id),
    is_futures bool not null,
    price_open numeric not null,
    price_close numeric,
    usd numeric not null,
    quantity numeric not null,
    profit numeric,
    roi numeric,
    status varchar(25),
    time_stamp timestamp without time zone,
    futures_side varchar(20) not null
);

CREATE TABLE public.instance_data (
                                      instance_id int references public.strategy_instances(id) on DELETE cascade,
                                      data jsonb
);

ALTER TABLE public.strategy_instances ADD COLUMN
    is_futures boolean not null DEFAULT false;

CREATE OR REPLACE VIEW public.strategy_instance_monitoring AS
SELECT
    instance.id, instance.strategy_id, instance.user_id, instance.pair, instance.bid,
    instance.time_frame, instance.status, strategy.name, (select sum(profit) as profit from public.trades where instance_id = instance.id),
       (select count(id)::decimal from public.trades where instance_id = instance.id and profit > 0)/(select CASE WHEN count(id) = 0 THEN 1 ELSE count(id)::decimal END from public.trades where instance_id = instance.id) as win_rate, instance.is_futures
FROM public.strategy_instances as instance
         LEFT JOIN strategy_info as strategy on instance.strategy_id = strategy.id;
