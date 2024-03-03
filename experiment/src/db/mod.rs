#[derive(Debug)]
pub struct User {
    pub id: i32,
    pub email: String,
    pub password: String,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug)]
pub struct Sessions {
    pub id: i32,
    pub user_id: i32,
    pub token: String,
    pub user_agent: String,
    pub created_at: String,
    pub updated_at: String,
    pub expires_at: String,
}

#[derive(Debug)]
pub struct JobApplications {
    pub id: i32,
    pub company: String,
    pub title: String,
    pub url: String,
    pub status: JobApplicationStatus,
    pub applied_at: String,
    pub created_at: String,
    pub updated_at: String,
    pub user_id: i32,
}

#[derive(Debug)]
pub struct JobApplicationStatusHistory {
    pub id: i32,
    pub job_application_id: i32,
    pub status: JobApplicationStatus,
    pub created_at: String,
}

#[derive(Debug)]
pub struct JobApplicationNotes {
    pub id: i32,
    pub job_application_id: i32,
    pub note: String,
    pub created_at: String,
}

pub enum JobApplicationTimeline {
    Note,
    Status,
}

impl JobApplicationTimeline {
    pub fn from_str(s: String) -> Self {
        match s.as_str() {
            "note" => Self::Note,
            "status" => Self::Status,
            _ => Self::Note,
        }
    }

    pub fn to_string(&self) -> String {
        match self {
            Self::Note => "note".to_string(),
            Self::Status => "status".to_string(),
        }
    }
}

#[derive(Debug)]
pub enum JobApplicationStatus {
    Accepted,
    Applied,
    Cancelled,
    Closed,
    Declined,
    Interviewing,
    Offered,
    Rejected,
    Watching,
    Withdrawn,
}

impl JobApplicationStatus {
    pub fn from_str(s: String) -> Self {
        match s.as_str() {
            "accepted" => Self::Accepted,
            "applied" => Self::Applied,
            "cancelled" => Self::Cancelled,
            "closed" => Self::Closed,
            "declined" => Self::Declined,
            "interviewing" => Self::Interviewing,
            "offered" => Self::Offered,
            "rejected" => Self::Rejected,
            "watching" => Self::Watching,
            "withdrawn" => Self::Withdrawn,
            _ => Self::Applied,
        }
    }

    pub fn to_string(&self) -> String {
        match self {
            Self::Accepted => "accepted".to_string(),
            Self::Applied => "applied".to_string(),
            Self::Cancelled => "cancelled".to_string(),
            Self::Closed => "closed".to_string(),
            Self::Declined => "declined".to_string(),
            Self::Interviewing => "interviewing".to_string(),
            Self::Offered => "offered".to_string(),
            Self::Rejected => "rejected".to_string(),
            Self::Watching => "watching".to_string(),
            Self::Withdrawn => "withdrawn".to_string(),
        }
    }

    pub fn pretty(&self) -> String {
        match self {
            Self::Accepted => "Accepted".to_string(),
            Self::Applied => "Applied".to_string(),
            Self::Cancelled => "Cancelled".to_string(),
            Self::Closed => "Closed".to_string(),
            Self::Declined => "Declined".to_string(),
            Self::Interviewing => "Interviewing".to_string(),
            Self::Offered => "Offered".to_string(),
            Self::Rejected => "Rejected".to_string(),
            Self::Watching => "Watching".to_string(),
            Self::Withdrawn => "Withdrawn".to_string(),
        }
    }
}
