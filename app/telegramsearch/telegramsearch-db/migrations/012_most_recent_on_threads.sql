---- tern: disable-tx ----

ALTER TABLE chatthreads
    ADD column most_recent_message_at TIMESTAMP WITH TIME ZONE DEFAULT '2011-05-19T09:45:17';
