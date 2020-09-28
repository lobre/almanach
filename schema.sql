-- Table events

DROP TABLE IF EXISTS events CASCADE;

CREATE TABLE events
(
    id INT GENERATED BY DEFAULT AS IDENTITY,
    name VARCHAR NOT NULL,
    date date NOT NULL,
    PRIMARY KEY (id)
);

-- Table subscriptions

DROP TABLE IF EXISTS subscriptions;

CREATE TABLE subscriptions
(
    event_id INT NOT NULL REFERENCES events(id),
    subscriber VARCHAR NOT NULL,
    here BOOLEAN NOT NULL,
    comment VARCHAR,
    PRIMARY KEY (event_id, subscriber)
);

-- Fake initial data

INSERT INTO events (id, name, date) VALUES (1, 'repet', '2020-09-26'); 
INSERT INTO events (id, name, date) VALUES (2, 'repet', '2020-11-14'); 
INSERT INTO events (id, name, date) VALUES (3, 'repet', '2020-11-28'); 
INSERT INTO events (id, name, date) VALUES (4, 'grande margot', '2020-10-11'); 
INSERT INTO events (id, name, date) VALUES (5, 'chandieux', '2021-03-07'); 

INSERT INTO subscriptions (event_id, subscriber, here, comment) VALUES (1, 'loric', true, 'no comment');
INSERT INTO subscriptions (event_id, subscriber, here, comment) VALUES (2, 'loric', false, 'still not sure 100%');
INSERT INTO subscriptions (event_id, subscriber, here, comment) VALUES (3, 'loric', false, 'no comment');
INSERT INTO subscriptions (event_id, subscriber, here, comment) VALUES (4, 'loric', false, '');
INSERT INTO subscriptions (event_id, subscriber, here, comment) VALUES (5, 'loric', true, '');

INSERT INTO subscriptions (event_id, subscriber, here, comment) VALUES (1, 'marco', true, 'will be late');
INSERT INTO subscriptions (event_id, subscriber, here, comment) VALUES (2, 'marco', false, 'need to garder my dog');
INSERT INTO subscriptions (event_id, subscriber, here, comment) VALUES (3, 'marco', true, 'we can repete in my swimming pool');
INSERT INTO subscriptions (event_id, subscriber, here, comment) VALUES (4, 'marco', true, '');
INSERT INTO subscriptions (event_id, subscriber, here, comment) VALUES (5, 'marco', true, 'so happy that we accepted');
