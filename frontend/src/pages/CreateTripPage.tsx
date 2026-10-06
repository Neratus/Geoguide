import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { createTrip } from "../api/trips";

export default function CreateTripPage() {
    const navigate = useNavigate();
    const [title, setTitle] = useState("");
    const [startDate, setStartDate] = useState("");
    const [endDate, setEndDate] = useState("");
    const [budget, setBudget] = useState("");
    const [notes, setNotes] = useState("");
    const [error, setError] = useState("");

    const today = new Date().toISOString().slice(0, 10);
    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (startDate < today) {
            setError("Start date cannot be in the past");
            return;
        }
        if (endDate < startDate) {
            setError("End date must be after start date");
            return;
        }
        try {
            const res = await createTrip({ title, startDate, endDate, budget: budget ? parseFloat(budget) : undefined, notes });
            navigate(`/trips/${res.data.id}`);
        } catch (err: any) {
            setError(err.response?.data?.error || "Failed to create trip");
        }
    };

    return (
        <div className="max-w-xl mx-auto">
            <h1 className="text-heading">Create New Trip</h1>
            <form onSubmit={handleSubmit} className="space-y-4 mt-4">
                <input type="text" placeholder="Trip title" value={title} onChange={e => setTitle(e.target.value)} required className="w-full" />
                <div className="grid grid-cols-2 gap-2">
                    <input type="date" value={startDate} onChange={e => setStartDate(e.target.value)} required className="w-full" />
                    <input type="date" value={endDate} onChange={e => setEndDate(e.target.value)} required className="w-full" />
                </div>
                <input type="number" placeholder="Budget (optional)" value={budget} onChange={e => setBudget(e.target.value)} className="w-full" />
                <textarea placeholder="Notes (optional)" rows={4} value={notes} onChange={e => setNotes(e.target.value)} className="w-full" />
                {error && <div className="text-accent">{error}</div>}
                <button type="submit" className="btn-primary">Create Trip</button>
            </form>
        </div>
    );
}