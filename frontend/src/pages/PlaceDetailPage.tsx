import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { getPlace } from "../api/catalog";
import { getFavourites, addFavourite, removeFavourite } from "../api/favourites";
import type { PlaceDetail } from "../types/api";
import { useAuth } from "../contexts/AuthContext";

export default function PlaceDetailPage() {
    const { id } = useParams<{ id: string }>();
    const [place, setPlace] = useState<PlaceDetail | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const { token } = useAuth();
    const [isFavourite, setIsFavourite] = useState(false);
    const [favLoading, setFavLoading] = useState(false);

    useEffect(() => {
        if (!id) return;
        const fetchPlace = async () => {
            try {
                const response = await getPlace(id);
                setPlace(response.data);
                setError(null);
            } catch (err: any) {
                console.error("Error fetching place:", err);
                setError(err.response?.data?.error || "Failed to load place");
            } finally {
                setLoading(false);
            }
        };
        fetchPlace();
    }, [id]);

    useEffect(() => {
        if (!id || !token) return;
        const checkFavourite = async () => {
            try {
                const favs = await getFavourites({ limit: 1000 });
                setIsFavourite(favs.some(f => f.id === id));
            } catch (err) {
                console.error("Failed to check favourite status", err);
            }
        };
        checkFavourite();
    }, [id, token]);

    const toggleFavourite = async () => {
        if (!token) {
            alert("Please login to add favourites");
            return;
        }
        setFavLoading(true);
        try {
            if (isFavourite) {
                await removeFavourite(id!);
                setIsFavourite(false);
            } else {
                await addFavourite(id!);
                setIsFavourite(true);
            }
        } catch (err: any) {
            alert(err.response?.data?.error || "Operation failed");
        } finally {
            setFavLoading(false);
        }
    };

    if (loading) return <div className="text-center p-8">Loading place...</div>;
    if (error) return <div className="text-accent text-center p-8">Error: {error}</div>;
    if (!place) return <div className="text-center p-8">Place not found</div>;

    return (
        <div>
            <h1 className="text-heading">{place.name}</h1>
            <div className="flex gap-2 my-2 flex-wrap items-center">
                <span className="bg-gray-200 px-2 py-1 rounded text-sm">{place.category}</span>
                {place.avgRating ? (
                    <span className="bg-yellow-100 px-2 py-1 rounded text-sm">⭐ {place.avgRating.toFixed(1)}</span>
                ) : null}
                {token && (
                    <button
                        onClick={toggleFavourite}
                        disabled={favLoading}
                        className={`px-4 py-1 rounded text-sm ${isFavourite ? "bg-red-600 hover:bg-red-700" : "bg-green-600 hover:bg-green-700"
                            } text-white transition`}
                    >
                        {favLoading ? "..." : isFavourite ? "! Remove from Favourites" : "+ Add to Favourites"}
                    </button>
                )}
            </div>
            <p className="text-secondary">{place.description}</p>

            <div className="grid md:grid-cols-2 gap-4 mt-4">
                <div>
                    <p><span className="font-medium">Address:</span> {place.address || "—"}</p>
                    <p><span className="font-medium">Opening hours:</span> {place.openingHours || "—"}</p>
                    <p><span className="font-medium">Price:</span> {place.priceInfo || "—"}</p>
                    <p><span className="font-medium">Avg. visit duration:</span> {place.avgVisitDurationMin || "—"} min</p>
                    {place.contactPhone && <p><span className="font-medium">Phone:</span> {place.contactPhone}</p>}
                    {place.website && (
                        <p>
                            <span className="font-medium">Website:</span>{" "}
                            <a href={place.website} target="_blank" rel="noopener noreferrer" className="text-primary">
                                {place.website}
                            </a>
                        </p>
                    )}
                </div>
                {place.coordinates && (
                    <div>
                        <p><span className="font-medium">Coordinates:</span> {place.coordinates}</p>
                    </div>
                )}
            </div>

            {token && (
                <Link to={`/places/${id}/reviews/new`} className="btn-primary mt-4 inline-block">
                    Write a Review
                </Link>
            )}

            <div className="mt-6">
                <h2 className="font-semibold">Reviews</h2>
                {place.reviews && place.reviews.length > 0 ? (
                    <div className="space-y-4 mt-2">
                        {place.reviews.map((rev) => (
                            <div key={rev.id} className="border p-3 rounded">
                                <div className="flex justify-between items-start">
                                    <div>
                                        <span className="font-medium">{rev.username || "Anonymous"}</span>
                                        <span className="ml-2 text-yellow-600">⭐ {rev.rating}</span>
                                    </div>
                                    <span className="text-secondary text-sm">{rev.visitDate}</span>
                                </div>
                                <p className="mt-1">{rev.comment}</p>
                                {rev.isApproved === false && <p className="text-accent text-sm mt-1"> Pending moderation</p>}
                            </div>
                        ))}
                    </div>
                ) : (
                    <p>No reviews yet.</p>
                )}
            </div>
        </div>
    );
}