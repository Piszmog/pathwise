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

-- Already has job apps
INSERT INTO users (
	email,
	password
)
VALUES (
	'user3@email.com',
	'$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm'
);

INSERT INTO job_applications (company, title, url, applied_at, user_id) 
SELECT 
	'Company A', 
	'Title A', 
	'http://companyA/titleA',
	datetime('now', '-2 days'),
	id
FROM users where email = 'user3@email.com';

INSERT INTO job_application_status_histories (
	job_application_id
) VALUES (
	last_insert_rowid()
);

INSERT INTO job_application_stats (total_applications, total_companies, total_applied, user_id)
SELECT 1, 1, 1, id FROM users where email = 'user3@email.com';

-- Additional test users for new tests
INSERT INTO users (
	email,
	password
)
VALUES (
	'user4@email.com',
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
	'user5@email.com',
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
	'user6@email.com',
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
	'user7@email.com',
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
	'user8@email.com',
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
	'user9@email.com',
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
	'user10@email.com',
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
	'user11@email.com',
	'$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm'
);

INSERT INTO job_application_stats (
	user_id
)
VALUES (
	last_insert_rowid()
);

-- Export test users
INSERT INTO users (
	email,
	password
)
VALUES (
	'export_user1@email.com',
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
	'export_user2@email.com',
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
	'export_user3@email.com',
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
	'export_user4@email.com',
	'$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm'
);

INSERT INTO job_application_stats (
	user_id
)
VALUES (
	last_insert_rowid()
);

-- Settings test users
INSERT INTO users (
	email,
	password
)
VALUES (
	'settings_user1@email.com',
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
	'settings_user2@email.com',
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
	'settings_user3@email.com',
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
	'settings_user4@email.com',
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
	'settings_user5@email.com',
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
	'settings_user6@email.com',
	'$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm'
);

INSERT INTO job_application_stats (
	user_id
)
VALUES (
	last_insert_rowid()
);

-- Salary test users
INSERT INTO users (
	email,
	password
)
VALUES (
	'salary_user1@email.com',
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
	'salary_user2@email.com',
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
	'salary_user3@email.com',
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
	'salary_user4@email.com',
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
	'salary_user5@email.com',
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
	'salary_user6@email.com',
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
	'salary_user7@email.com',
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
	'salary_user8@email.com',
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
	'salary_user9@email.com',
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
	'salary_user10@email.com',
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
	'salary_user11@email.com',
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
	'salary_user12@email.com',
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
	'salary_user13@email.com',
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
	'salary_user14@email.com',
	'$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm'
);

INSERT INTO job_application_stats (
	user_id
)
VALUES (
	last_insert_rowid()
);

-- Pagination test users
INSERT INTO users (
	email,
	password
)
VALUES (
	'pagination_user1@email.com',
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
	'pagination_user2@email.com',
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
	'pagination_user3@email.com',
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
	'pagination_user4@email.com',
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
	'pagination_user5@email.com',
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
	'pagination_user6@email.com',
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
	'pagination_user7@email.com',
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
	'pagination_user8@email.com',
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
	'pagination_user9@email.com',
	'$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm'
);

INSERT INTO job_application_stats (
	user_id
)
VALUES (
	last_insert_rowid()
);
