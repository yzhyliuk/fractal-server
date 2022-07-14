ALTER TABLE public.strategy_instances ADD COLUMN
    leverage integer;

ALTER TABLE public.trades ADD COLUMN
    leverage integer;

INSERT INTO public.strategy_fields (strategy_id, name, display_name, description, min, max, default_value, type, ui_type, dataset)
    VALUES (0, 'leverage','Adjust Leverage', 'Set leverage for the asset: min = 1', 1, 50, '10','number','input', NULL);

ALTER TABLE public.strategy_fields ADD COLUMN
    futuresOnly bool DEFAULT false;

UPDATE public.strategy_fields SET futuresOnly = true WHERE name = 'leverage';

DELETE FROM public.strategy_fields WHERE name = 'isFutures';