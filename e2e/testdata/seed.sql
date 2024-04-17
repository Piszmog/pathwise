INSERT INTO users (
	email,
	password
)
VALUES (
	'user1@email.com',
	'$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm'
);

INSERT INTO job_application_stats (
	user_id
)
VALUES (
	last_insert_rowid()
);

INSERT INTO users (
	email,
	password
)
VALUES (
	'user2@email.com',
	'$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm'
);

INSERT INTO job_application_stats (
	user_id
)
VALUES (
	last_insert_rowid()
);
