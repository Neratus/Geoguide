import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { getCity, getPlacesByCity, getCountry } from "../api/catalog";
import type { CityDetail, Place, Country } from "../types/api";

export default function CityDetailPage() {
    const { id } = useParams<{ id: string }>();
    const [city, setCity] = useState<CityDetail | null>(null);
    const [places, setPlaces] = useState<Place[]>([]);
    const [country, setCountry] = useState<Country | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!id) return;
        const fetchData = async () => {
            setLoading(true);
            try {
                const cityRes = await getCity(id);
                setCity(cityRes.data);
                if (cityRes.data.countryId) {
                    try {
                        const countryRes = await getCountry(cityRes.data.countryId);
                        setCountry(countryRes.data);
                    } catch (err) {
                        console.error("Failed to load country", err);
                    }
                }
                const placesRes = await getPlacesByCity(id, { limit: 100 });
                setPlaces(placesRes.data);
                setError(null);
            } catch (err: any) {
                setError(err.response?.data?.error || "Failed to load city details");
            } finally {
                setLoading(false);
            }
        };
        fetchData();
    }, [id]);

    if (loading) return <div className="text-center p-8">Loading...</div>;
    if (error) return <div className="text-accent text-center p-8">{error}</div>;
    if (!city) return <div className="text-center p-8">City not found</div>;

    return (
        <div>
            <h1 className="text-heading">{city.name}</h1>
            {city.imageUrl && <img src={city.imageUrl} alt={city.name} className="w-full h-64 object-cover rounded my-4" />}
            <div className="grid md:grid-cols-2 gap-4">
                <div>
                    <p>
                        <span className="font-medium">Country:</span>{" "}
                        <Link to={`/countries/${city.countryId}`} className="text-primary">
                            {country?.name || city.countryId}
                        </Link>
                    </p>
                    <p><span className="font-medium">Population:</span> {city.population?.toLocaleString()}</p>
                    <p><span className="font-medium">Timezone:</span> {city.timezone}</p>
                    {city.isCapital && <p><span className="font-medium">Capital</span></p>}
                </div>
                <div>
                    <p><span className="font-medium">Description:</span> {city.description}</p>
                    {city.travelTips && <p><span className="font-medium">Travel tips:</span> {city.travelTips}</p>}
                </div>
            </div>

            <div className="mt-6">
                <h2 className="font-semibold">Attractions & Places</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-2">
                    {places.map((place) => (
                        <Link key={place.id} to={`/places/${place.id}`} className="block border rounded p-3 hover:shadow">
                            <h3 className="font-medium">{place.name}</h3>
                            <p className="text-secondary text-sm">{place.category} • ⭐ {place.avgRating?.toFixed(1) || "N/A"} ({place.reviewsCount} reviews)</p>
                            {place.address && <p className="text-secondary text-sm">{place.address}</p>}
                        </Link>
                    ))}
                </div>
            </div>
        </div>
    );
}