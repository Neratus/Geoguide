import { useEffect, useState } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import { getTrip, updateTrip, addPlaceToTrip, removePlaceFromTrip } from "../api/trips";
import { searchCities, getPlacesByCity } from "../api/catalog";
import { getFavourites } from "../api/favourites";
import type { TripDetail, City, Place } from "../types/api";

export default function TripDetailPage() {
    const { id } = useParams();
    const navigate = useNavigate();
    const [trip, setTrip] = useState<TripDetail | null>(null);
    const [loading, setLoading] = useState(true);
    const [editing, setEditing] = useState(false);
    const [formData, setFormData] = useState({ title: "", startDate: "", endDate: "", budget: 0, notes: "" });

    const [cities, setCities] = useState<City[]>([]);
    const [selectedCityId, setSelectedCityId] = useState("");
    const [places, setPlaces] = useState<Place[]>([]);
    const [selectedPlaceId, setSelectedPlaceId] = useState("");
    const [dayNumber, setDayNumber] = useState(1);
    const [durationMin, setDurationMin] = useState(60);
    const [arrivalTime, setArrivalTime] = useState("");
    const [placeNotes, setPlaceNotes] = useState("");
    const [message, setMessage] = useState("");

    const [favourites, setFavourites] = useState<Place[]>([]);
    const [selectedFavPlaceId, setSelectedFavPlaceId] = useState("");

    const fetchTrip = async () => {
        if (!id) return;
        try {
            const res = await getTrip(id);
            setTrip(res.data);
            setFormData({
                title: res.data.title,
                startDate: res.data.startDate,
                endDate: res.data.endDate,
                budget: res.data.budget || 0,
                notes: res.data.notes || "",
            });
        } catch (err: any) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    const fetchCities = async () => {
        try {
            const res = await searchCities("", { limit: 200 });
            setCities(res.data);
        } catch (err) {
            console.error("Failed to load cities", err);
        }
    };

    const loadFavourites = async () => {
        try {
            const data = await getFavourites({ limit: 200 });
            setFavourites(data);
        } catch (err) {
            console.error("Failed to load favourites", err);
        }
    };

    useEffect(() => {
        fetchTrip();
        fetchCities();
        loadFavourites();
    }, [id]);

    useEffect(() => {
        if (!selectedCityId) {
            setPlaces([]);
            setSelectedPlaceId("");
            return;
        }
        const loadPlaces = async () => {
            try {
                const res = await getPlacesByCity(selectedCityId, { limit: 100 });
                setPlaces(res.data);
            } catch (err) {
                console.error("Failed to load places", err);
            }
        };
        loadPlaces();
    }, [selectedCityId]);

    const handleUpdate = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await updateTrip(id!, formData);
            setEditing(false);
            fetchTrip();
            setMessage("Trip updated");
            setTimeout(() => setMessage(""), 3000);
        } catch (err: any) {
            alert(err.response?.data?.error);
        }
    };

    const handleAddPlace = async () => {
        if (!selectedPlaceId) {
            alert("Please select a place");
            return;
        }
        const payload: any = {
            placeId: selectedPlaceId,
            dayNumber,
            durationMin,
        };
        if (arrivalTime) payload.arrivalTime = arrivalTime;
        if (placeNotes) payload.notes = placeNotes;

        try {
            await addPlaceToTrip(id!, payload);
            fetchTrip();
            setMessage("Place added");
            setTimeout(() => setMessage(""), 3000);
            setSelectedCityId("");
            setSelectedPlaceId("");
            setDayNumber(1);
            setDurationMin(60);
            setArrivalTime("");
            setPlaceNotes("");
        } catch (err: any) {
            alert(err.response?.data?.error);
        }
    };

    const handleAddFavPlace = async () => {
        if (!selectedFavPlaceId) {
            alert("Select a place from favourites");
            return;
        }
        const payload: any = {
            placeId: selectedFavPlaceId,
            dayNumber,
            durationMin,
        };
        if (arrivalTime) payload.arrivalTime = arrivalTime;
        if (placeNotes) payload.notes = placeNotes;

        try {
            await addPlaceToTrip(id!, payload);
            fetchTrip();
            setMessage("Place added from favourites");
            setTimeout(() => setMessage(""), 3000);
            setSelectedFavPlaceId("");
        } catch (err: any) {
            alert(err.response?.data?.error);
        }
    };

    const handleRemovePlace = async (placeId: string) => {
        if (!confirm("Remove this place from trip?")) return;
        try {
            await removePlaceFromTrip(id!, placeId);
            fetchTrip();
        } catch (err: any) {
            alert(err.response?.data?.error);
        }
    };

    if (loading) return <div>Loading trip...</div>;
    if (!trip) return <div>Trip not found</div>;

    return (
        <div>
            <div className="flex justify-between items-start">
                {!editing ? (
                    <div>
                        <h1 className="text-heading">{trip.title}</h1>
                        <p>
                            {trip.startDate} → {trip.endDate} | Budget: {trip.budget} | Status: {trip.status}
                        </p>
                        <p className="mt-2">{trip.notes}</p>
                        <button onClick={() => setEditing(true)} className="btn-secondary mt-2">
                            Edit
                        </button>
                    </div>
                ) : (
                    <form onSubmit={handleUpdate} className="space-y-2 flex-grow">
                        <input
                            value={formData.title}
                            onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                            className="w-full"
                        />
                        <div className="flex gap-2">
                            <input
                                type="date"
                                value={formData.startDate}
                                onChange={(e) => setFormData({ ...formData, startDate: e.target.value })}
                            />
                            <input
                                type="date"
                                value={formData.endDate}
                                onChange={(e) => setFormData({ ...formData, endDate: e.target.value })}
                            />
                        </div>
                        <input
                            type="number"
                            value={formData.budget}
                            onChange={(e) => setFormData({ ...formData, budget: parseFloat(e.target.value) })}
                            placeholder="Budget"
                        />
                        <textarea
                            value={formData.notes}
                            onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                            rows={3}
                        />
                        <div className="flex gap-2">
                            <button type="submit" className="btn-primary">
                                Save
                            </button>
                            <button type="button" onClick={() => setEditing(false)} className="btn-secondary">
                                Cancel
                            </button>
                        </div>
                    </form>
                )}
            </div>

            {message && <div className="mt-2 p-2 bg-green-100 text-green-800 rounded">{message}</div>}

            <hr className="my-6" />

            <h2 className="font-semibold">Add a Place</h2>
            <div className="flex flex-col gap-2 mt-2">
                <select
                    value={selectedCityId}
                    onChange={(e) => setSelectedCityId(e.target.value)}
                    className="border p-2"
                >
                    <option value="">Select city</option>
                    {cities.map((city) => (
                        <option key={city.id} value={city.id}>
                            {city.name}
                        </option>
                    ))}
                </select>
                <select
                    value={selectedPlaceId}
                    onChange={(e) => setSelectedPlaceId(e.target.value)}
                    disabled={!selectedCityId}
                    className="border p-2"
                >
                    <option value="">Select place</option>
                    {places.map((place) => (
                        <option key={place.id} value={place.id}>
                            {place.name} ({place.category})
                        </option>
                    ))}
                </select>
                <div className="flex flex-wrap gap-2">
                    <input
                        type="number"
                        placeholder="Day"
                        value={dayNumber}
                        onChange={(e) => setDayNumber(parseInt(e.target.value))}
                        className="w-20 border p-2"
                    />
                    <input
                        type="number"
                        placeholder="Duration (min)"
                        value={durationMin}
                        onChange={(e) => setDurationMin(parseInt(e.target.value))}
                        className="w-28 border p-2"
                    />
                    <input
                        type="text"
                        placeholder="Notes (optional)"
                        value={placeNotes}
                        onChange={(e) => setPlaceNotes(e.target.value)}
                        className="flex-1 border p-2"
                    />
                    <button onClick={handleAddPlace} className="btn-primary">
                        Add
                    </button>
                </div>
            </div>

            <div className="mt-4">
                <select
                    value={selectedFavPlaceId}
                    onChange={(e) => setSelectedFavPlaceId(e.target.value)}
                    className="border p-2 w-full"
                >
                    <option value="">-- Add from favourites --</option>
                    {favourites.map((fav) => (
                        <option key={fav.id} value={fav.id}>
                            {fav.name} ({fav.cityName})
                        </option>
                    ))}
                </select>
                <button onClick={handleAddFavPlace} className="btn-primary mt-1 w-full">
                    Add from Favourites
                </button>
            </div>

            <h2 className="font-semibold mt-6">Places in Trip</h2>
            <div className="space-y-2 mt-2">
                {trip.places?.map((p) => (
                    <div key={p.placeId} className="border p-3 rounded flex justify-between items-center">
                        <div>
                            <span className="font-medium">Day {p.dayNumber}</span> –{" "}
                            <Link to={`/places/${p.placeId}`} className="text-primary">
                                {p.placeId}
                            </Link>
                            {p.durationMin && <span> • {p.durationMin} min</span>}
                            {p.arrivalTime && (
                                <span>
                                    {" "}
                                    • arrival{" "}
                                    {new Date(p.arrivalTime).toLocaleTimeString([], {
                                        hour: "2-digit",
                                        minute: "2-digit",
                                    })}
                                </span>
                            )}
                            <div className="text-secondary text-sm">{p.notes}</div>
                        </div>
                        <button onClick={() => handleRemovePlace(p.placeId)} className="text-accent">
                            Remove
                        </button>
                    </div>
                ))}
            </div>
        </div>
    );
}