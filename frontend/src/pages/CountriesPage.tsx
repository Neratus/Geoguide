import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getCountries } from "../api/catalog";
import type { Country } from "../types/api";

export default function CountriesPage() {
    const [countries, setCountries] = useState<Country[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [searchInput, setSearchInput] = useState("");
    const [search, setSearch] = useState("");
    const [page, setPage] = useState(0);
    const limit = 20;

    const fetchCountries = async () => {
        setLoading(true);
        try {
            const response = await getCountries({ limit, offset: page * limit, search: search || undefined });
            setCountries(response.data);
            setError(null);
        } catch (err: any) {
            setError(err.response?.data?.error || "Failed to load countries");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchCountries();
    }, [search, page]);

    const handleSearchSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        setSearch(searchInput);
        setPage(0);
    };

    if (loading && page === 0) return <div className="text-center p-8">Loading countries...</div>;
    if (error) return <div className="text-accent text-center p-8">{error}</div>;

    return (
        <div>
            <h1 className="text-heading mb-6">Countries of the World</h1>
            <form onSubmit={handleSearchSubmit} className="mb-6 flex gap-2">
                <input
                    type="text"
                    placeholder="Search by name..."
                    value={searchInput}
                    onChange={(e) => setSearchInput(e.target.value)}
                    className="flex-grow"
                />
                <button type="submit" className="btn-primary">Search</button>
            </form>
            {countries.length === 0 ? (
                <div className="text-center">No countries found.</div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {countries.map((country) => (
                        <Link key={country.id} to={`/countries/${country.id}`} className="block border rounded p-4 hover:shadow transition">
                            {country.imageUrl && <img src={country.imageUrl} alt={country.name} className="w-full h-40 object-cover rounded mb-2" />}
                            <h2 className="font-semibold text-lg">{country.name}</h2>
                            {country.capital && <p className="text-secondary">Capital: {country.capital}</p>}
                            {country.population && <p className="text-secondary">Population: {country.population.toLocaleString()}</p>}
                        </Link>
                    ))}
                </div>
            )}

            <div className="flex justify-between items-center mt-6">
                <button
                    onClick={() => setPage((p) => Math.max(0, p - 1))}
                    disabled={page === 0}
                    className="btn-secondary disabled:opacity-50"
                >
                    Previous
                </button>
                <span>Page {page + 1}</span>
                <button
                    onClick={() => setPage((p) => p + 1)}
                    disabled={countries.length < limit}
                    className="btn-secondary disabled:opacity-50"
                >
                    Next
                </button>
            </div>
        </div>
    );
}