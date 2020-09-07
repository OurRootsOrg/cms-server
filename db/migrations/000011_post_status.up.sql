update post set body = jsonb_set(body, '{postStatus}', body->'recordsStatus');
update post set body = jsonb_set(body, '{recordsStatus}', '""');
update post set body = jsonb_set(body, '{imagesStatus}', '""');