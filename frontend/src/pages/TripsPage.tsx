import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getUserTrips, deleteTrip } from "../api/trips";
import type { Trip } from "../types/api";

export default function TripsPage() {
    const [trips, setTrips] = useState<Trip[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");

    const fetchTrips = async () => {
        try {
            const res = await getUserTrips();
            setTrips(res.data);
        } catch (err: any) {
            setError(err.response?.data?.error || "Failed to load trips");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { fetchTrips(); }, []);

    const handleDelete = async (id: string) => {
        if (!confirm("Delete this trip?")) return;
        try {
            await deleteTrip(id);
            setTrips(trips.filter(t => t.id !== id));
        } catch (err: any) {
            alert(err.response?.data?.error);
        }
    };

    if (loading) return <div>Loading trips...</div>;
    if (error) return <div className="text-accent">{error}</div>;

    return (
        <div>
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-heading">My Trips</h1>
                <Link to="/trips/new" className="btn-primary">+ New Trip</Link>
            </div>
            {trips.length === 0 ? (
                <p>No trips yet. Create your first trip!</p>
            ) : (
                <div className="space-y-4">
                    {trips.map(trip => (
                        <div key={trip.id} className="border p-4 rounded flex justify-between items-center">
                            <Link to={`/trips/${trip.id}`} className="flex-grow">
                                <h3 className="font-semibold">{trip.title}</h3>
                                <p className="text-secondary text-sm">{trip.startDate} → {trip.endDate} • {trip.status}</p>
                            </Link>
                            <button onClick={() => handleDelete(trip.id)} className="text-accent hover:underline">Delete</button>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}