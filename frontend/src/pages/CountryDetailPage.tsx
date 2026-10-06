import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { getCountry, getCitiesByCountry } from "../api/catalog";
import type { CountryDetail, City } from "../types/api";

export default function CountryDetailPage() {
    const { id } = useParams<{ id: string }>();
    const [country, setCountry] = useState<CountryDetail | null>(null);
    const [cities, setCities] = useState<City[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!id) return;
        const fetchData = async () => {
            setLoading(true);
            try {
                const [countryRes, citiesRes] = await Promise.all([
                    getCountry(id),
                    getCitiesByCountry(id, { limit: 50 }),
                ]);
                setCountry(countryRes.data);
                setCities(citiesRes.data);
                setError(null);
            } catch (err: any) {
                setError(err.response?.data?.error || "Failed to load country details");
            } finally {
                setLoading(false);
            }
        };
        fetchData();
    }, [id]);

    if (loading) return <div className="text-center p-8">Loading...</div>;
    if (error) return <div className="text-accent text-center p-8">{error}</div>;
    if (!country) return <div className="text-center p-8">Country not found</div>;

    return (
        <div>
            {country.imageUrl && <img src={country.imageUrl} alt={country.name} className="w-full h-84 object-cover rounded mb-4" />}
            <h1 className="text-heading">{country.name}</h1>
            <div className="grid md:grid-cols-2 gap-6 mt-4">
                <div>
                    <h2 className="font-semibold">General Information</h2>
                    <ul className="mt-2 space-y-1">
                        <li><span className="font-medium">Capital:</span> {country.capitalName || "—"}</li>
                        <li><span className="font-medium">Area:</span> {country.area?.toLocaleString()} km²</li>
                        <li><span className="font-medium">Population:</span> {country.population?.toLocaleString()}</li>
                        <li><span className="font-medium">GDP:</span> {country.gdp?.toLocaleString()} USD</li>
                        <li><span className="font-medium">Currency:</span> {country.currency}</li>
                        <li><span className="font-medium">Language:</span> {country.language}</li>
                        <li><span className="font-medium">Phone code:</span> {country.phoneCode}</li>
                        <li><span className="font-medium">Religion:</span> {country.religion}</li>
                        <li><span className="font-medium">Best season:</span> {country.bestSeason}</li>
                    </ul>
                </div>
                <div>
                    <h2 className="font-semibold">Travel Info</h2>
                    <div className="mt-2 space-y-2">
                        <p><span className="font-medium">Description:</span> {country.description || "—"}</p>
                        <p><span className="font-medium">Visa requirements:</span> {country.visaRequirements || "—"}</p>
                        <p><span className="font-medium">Safety tips:</span> {country.safetyTips || "—"}</p>
                    </div>
                </div>
            </div>

            {country.holidays && country.holidays.length > 0 && (
                <div className="mt-6">
                    <h2 className="font-semibold">Public Holidays</h2>
                    <ul className="mt-2 list-disc list-inside">
                        {country.holidays.map((h) => (
                            <li key={h.id}>{h.name} – {h.date} {h.isNational && "(National)"}</li>
                        ))}
                    </ul>
                </div>
            )}

            <div className="mt-6">
                <h2 className="font-semibold">Major Cities</h2>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 mt-2">
                    {cities.map((city) => (
                        <Link key={city.id} to={`/cities/${city.id}`} className="block p-2 border rounded hover:bg-gray-50">
                            {city.name} {city.isCapital && "(Capital)"}
                        </Link>
                    ))}
                </div>
            </div>
        </div>
    );
}