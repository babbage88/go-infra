
-- name: GetLatestExternalAuthTokenByAppName :one
--SELECT et.id, et.user_id, et.external_app_id, ea.name 
--FROM external_auth_tokens et 
--LEFT JOIN public.external_integration_apps ea on et.external_app_id = ea.id
--WHERE et.user_id = $1 AND ea.name = $2
--ORDER BY et.created_at DESC
--LIMIT 1;

-- name: GetUserSecretsByUserId :many
SELECT
  auth_token_id,
  user_id,
  application_id,
  username,
  endpoint_url,
  email,
  application_name,
  token_created_at,
  expiration
FROM public.user_auth_app_mappings;
--WHERE username = 'devuser' and application_name = 'CloudflareDNS';

INSERT INTO public.user_role_mapping(user_id, role_id, enabled)
VALUES ('99d41ae8-09ac-4e1b-87a8-318cf03f6695', '10db2807-6c74-4dc0-aed7-c046c76bebdf', TRUE)
ON CONFLICT (user_id, role_id)
DO UPDATE SET enabled = TRUE;


