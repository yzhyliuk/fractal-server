INSERT INTO public.strategy_info (name, description)
VALUES ('Glide On Price', 'Simple strategy, that uses slope of price in current moment to follow the trend.');
INSERT INTO public.strategy_fields (strategy_id, name, display_name, description, min, max, default_value, type, ui_type, dataset)
VALUES (2,'volatilityTF', 'Volatility Time Frame', 'Number of timeframes for which volatility will be calculated',2,NULL,7,'number','input', NULL),
       (2,'slopeTF', 'Slope Time Frame', 'Number of timeframes for which slope will be calculated',2,NULL,2,'number','input',NULL),
       (2,'volatilityLimit', 'Volatility Limit', 'Min value of volatility for which strategy would place buy or sell orders. Calculates as HIGH\LOW on given timeframe',1,NULL, 1.001,'number','input',NULL);