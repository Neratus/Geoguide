import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getFavourites, removeFavourite } from "../api/favourites";
import type { Place } from "../types/api";

export default function FavouritesPage() {
    const [places, setPlaces] = useState<Place[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");

    const fetchFavourites = async () => {
        try {
            const data = await getFavourites({ limit: 100 });
            setPlaces(data);
        } catch (err: any) {
            setError(err.response?.data?.error || "Failed to load favourites");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchFavourites();
    }, []);

    const handleRemove = async (placeId: string) => {
        if (!confirm("Remove from favourites?")) return;
        try {
            await removeFavourite(placeId);
            setPlaces(places.filter(p => p.id !== placeId));
        } catch (err: any) {
            alert(err.response?.data?.error || "Failed to remove");
        }
    };

    if (loading) return <div className="text-center p-8">Loading favourites...</div>;
    if (error) return <div className="text-accent text-center p-8">{error}</div>;

    return (
        <div>
            <h1 className="text-heading">My Favourites</h1>
            {places.length === 0 ? (
                <p>You haven't added any favourite places yet.</p>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
                    {places.map(place => (
                        <div key={place.id} className="border rounded p-4 flex justify-between items-start">
                            <div>
                                <Link to={`/places/${place.id}`} className="font-semibold text-primary hover:underline">
                                    {place.name}
                                </Link>
                                <p className="text-secondary text-sm">{place.category}</p>
                                {place.cityName && <p className="text-secondary text-sm">{place.cityName}</p>}
                                {place.avgRating && place.avgRating > 0 && (
                                    <span className="text-yellow-600">⭐ {place.avgRating.toFixed(1)}</span>
                                )}                            </div>
                            <button
                                onClick={() => handleRemove(place.id)}
                                className="text-accent hover:underline"
                            >
                                Remove
                            </button>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}