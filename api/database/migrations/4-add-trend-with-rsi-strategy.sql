INSERT INTO public.strategy_info (name, description)
    VALUES ('Trend follow with RSI', 'Trend following strategy that buys and sells assets inside long term trend based on rsi indicator');

INSERT INTO public.strategy_fields (strategy_id, name, display_name, description, min, max, default_value, type, ui_type, dataset, futuresonly)
VALUES (3, 'rsiPeriod','RSI Period', 'Number of timeframes for RSI calculation', 2, NULL, 14, 'number','input',NULL, false),
       (3, 'trendMA', 'Trend period', 'Used for calculation of moving average to define current trend', 2, NULL, 100, 'number', 'input', NULL, false),
       (3, 'rsiOversoldLevel', 'Oversold level (RSI)', 'RSI limit for market to be considered oversold', 0, 49,30, 'number','input', NULL, false),
       (3, 'rsiOverboughtLevel', 'Overbought level (RSI)', 'RSI limit for market to be considered overbought',51,100,70, 'number','input',NULL, false);

CREATE OR REPLACE VIEW public.strategy_instance_monitoring AS
SELECT
    instance.id, instance.strategy_id, instance.user_id, instance.pair, instance.bid,
    instance.time_frame, instance.status, strategy.name, (select sum(profit) as profit from public.trades where instance_id = instance.id),
    (select count(id)::decimal from public.trades where instance_id = instance.id and profit > 0)/(select CASE WHEN count(id) = 0 THEN 1 ELSE count(id)::decimal END from public.trades where instance_id = instance.id) as win_rate, instance.is_futures,
       instance.leverage
FROM public.strategy_instances as instance
         LEFT JOIN strategy_info as strategy on instance.strategy_id = strategy.id;
