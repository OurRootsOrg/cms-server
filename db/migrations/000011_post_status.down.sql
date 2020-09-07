update post set body = jsonb_set(body, '{recordsStatus}', body->'postStatus');
update post set body = jsonb_set(body, '{postStatus}', '""');