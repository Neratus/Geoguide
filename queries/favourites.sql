-- name: AddFavorite :exec
INSERT INTO "UserFavoritePlace" (user_id, place_id)
VALUES ($1, $2)
ON CONFLICT (user_id, place_id) DO NOTHING;

-- name: RemoveFavorite :exec
DELETE FROM "UserFavoritePlace"
WHERE user_id = $1 AND place_id = $2;

-- name: GetFavoritesByUser :many
SELECT 
    p.id, p.name, p.category, p.description, p.coordinates, p.address, 
    p.opening_hours, p.price_info, p.avg_visit_duration_min, p.avg_rating, 
    p.reviews_count, p.contact_phone, p.website, p.city_id, p.district_id, p.image_id,
    c.name AS city_name, c.id AS city_id,
    -- добавим также created_at из избранного
    ufp.created_at AS favorited_at
FROM "UserFavoritePlace" ufp
JOIN "Place" p ON p.id = ufp.place_id
LEFT JOIN "City" c ON c.id = p.city_id
WHERE ufp.user_id = $1
ORDER BY ufp.created_at DESC
LIMIT $2 OFFSET $3;