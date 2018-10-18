--
-- PostgreSQL database dump
--

-- Dumped from database version 10.5 (Ubuntu 10.5-0ubuntu0.18.04)
-- Dumped by pg_dump version 10.5 (Ubuntu 10.5-0ubuntu0.18.04)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- Name: less_than(text, text); Type: FUNCTION; Schema: public; Owner: weather_api_user
--

CREATE FUNCTION public.less_than(a text, b text) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    local_a text := a;
    local_b text := b;
BEGIN
    RETURN local_a < local_b;
END;
$$;


ALTER FUNCTION public.less_than(a text, b text) OWNER TO weather_api_user;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: prognoze; Type: TABLE; Schema: public; Owner: weather_api_user
--

CREATE TABLE public.prognoze (
    utrka_id integer NOT NULL,
    ikona_stanja character varying(30),
    vrijeme_prognoze timestamp without time zone,
    kisa numeric,
    snijeg numeric,
    temperatura numeric,
    vlaznost numeric,
    brzina_vjetra numeric
);


ALTER TABLE public.prognoze OWNER TO weather_api_user;

--
-- Name: prognoze_utrka_id_seq; Type: SEQUENCE; Schema: public; Owner: weather_api_user
--

CREATE SEQUENCE public.prognoze_utrka_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.prognoze_utrka_id_seq OWNER TO weather_api_user;

--
-- Name: prognoze_utrka_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: weather_api_user
--

ALTER SEQUENCE public.prognoze_utrka_id_seq OWNED BY public.prognoze.utrka_id;


--
-- Name: utrke; Type: TABLE; Schema: public; Owner: weather_api_user
--

CREATE TABLE public.utrke (
    utrka_id integer NOT NULL,
    naziv character varying(100) NOT NULL,
    lat numeric NOT NULL,
    lon numeric NOT NULL,
    pocetak timestamp without time zone NOT NULL,
    kraj timestamp without time zone NOT NULL
);


ALTER TABLE public.utrke OWNER TO weather_api_user;

--
-- Name: utrke_utrka_id_seq; Type: SEQUENCE; Schema: public; Owner: weather_api_user
--

CREATE SEQUENCE public.utrke_utrka_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.utrke_utrka_id_seq OWNER TO weather_api_user;

--
-- Name: utrke_utrka_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: weather_api_user
--

ALTER SEQUENCE public.utrke_utrka_id_seq OWNED BY public.utrke.utrka_id;


--
-- Name: prognoze utrka_id; Type: DEFAULT; Schema: public; Owner: weather_api_user
--

ALTER TABLE ONLY public.prognoze ALTER COLUMN utrka_id SET DEFAULT nextval('public.prognoze_utrka_id_seq'::regclass);


--
-- Name: utrke utrka_id; Type: DEFAULT; Schema: public; Owner: weather_api_user
--

ALTER TABLE ONLY public.utrke ALTER COLUMN utrka_id SET DEFAULT nextval('public.utrke_utrka_id_seq'::regclass);


--
-- Data for Name: prognoze; Type: TABLE DATA; Schema: public; Owner: weather_api_user
--

COPY public.prognoze (utrka_id, ikona_stanja, vrijeme_prognoze, kisa, snijeg, temperatura, vlaznost, brzina_vjetra) FROM stdin;
173	vedro	2018-10-20 12:00:00	0	0	17.7	52	1.18
173	isprekidani oblaci	2018-10-20 15:00:00	0	0	16.09	54	1.18
173	slaba kiša	2018-10-20 18:00:00	0.0375	0	12.69	72	0.97
\.


--
-- Data for Name: utrke; Type: TABLE DATA; Schema: public; Owner: weather_api_user
--

COPY public.utrke (utrka_id, naziv, lat, lon, pocetak, kraj) FROM stdin;
173	SARAJEVO	43.866667	18.416667	2018-10-20 12:00:00	2018-10-20 18:00:00
\.


--
-- Name: prognoze_utrka_id_seq; Type: SEQUENCE SET; Schema: public; Owner: weather_api_user
--

SELECT pg_catalog.setval('public.prognoze_utrka_id_seq', 1, false);


--
-- Name: utrke_utrka_id_seq; Type: SEQUENCE SET; Schema: public; Owner: weather_api_user
--

SELECT pg_catalog.setval('public.utrke_utrka_id_seq', 174, true);


--
-- Name: utrke utrke_pkey; Type: CONSTRAINT; Schema: public; Owner: weather_api_user
--

ALTER TABLE ONLY public.utrke
    ADD CONSTRAINT utrke_pkey PRIMARY KEY (utrka_id);


--
-- Name: prognoze prognoze_utrka_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: weather_api_user
--

ALTER TABLE ONLY public.prognoze
    ADD CONSTRAINT prognoze_utrka_id_fkey FOREIGN KEY (utrka_id) REFERENCES public.utrke(utrka_id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

