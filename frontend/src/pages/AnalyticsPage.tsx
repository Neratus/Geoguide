import { useState } from "react";
import { generateReport } from "../api/analytics";
import type { ReportRequest, PopularPlace, UserActivity, TripStat } from "../types/api";

export default function AnalyticsPage() {
    const [reportType, setReportType] = useState<ReportRequest["reportType"]>("popular_places");
    const [format, setFormat] = useState<ReportRequest["format"]>("json");
    const [limit, setLimit] = useState(20);
    const [dateFrom, setDateFrom] = useState("");
    const [dateTo, setDateTo] = useState("");
    const [result, setResult] = useState<any>(null);
    const [loading, setLoading] = useState(false);

    const handleGenerate = async () => {
        setLoading(true);
        try {
            const data = await generateReport({ reportType, format, limit: 100 });
            console.log("Received report:", data);
            setResult(data);
        } catch (err: any) {
            alert(err.response?.data?.error);
        } finally {
            setLoading(false);
        }
    };

    const downloadCSV = () => {
        if (typeof result === "string") {
            const blob = new Blob([result], { type: "text/csv" });
            const url = URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = `report_${reportType}.csv`;
            a.click();
            URL.revokeObjectURL(url);
        }
    };

    return (
        <div className="max-w-4xl mx-auto">
            <h1 className="text-heading">Analytics Reports</h1>
            <div className="grid grid-cols-2 gap-4 mt-4">
                <div>
                    <label className="block font-medium">Report Type</label>
                    <select value={reportType} onChange={e => setReportType(e.target.value as any)} className="w-full">
                        <option value="popular_places">Popular Places</option>
                        <option value="user_activity">User Activity</option>
                        <option value="trip_statistics">Trip Statistics</option>
                    </select>
                </div>
                <div>
                    <label className="block font-medium">Format</label>
                    <select value={format} onChange={e => setFormat(e.target.value as any)} className="w-full">
                        <option value="json">JSON</option>
                        <option value="csv">CSV</option>
                    </select>
                </div>
                <div>
                    <label className="block font-medium">Limit</label>
                    <input type="number" min={1} max={1000} value={limit} onChange={e => setLimit(parseInt(e.target.value))} className="w-full" />
                </div>
                <div>
                    <label className="block font-medium">Date From (optional)</label>
                    <input type="date" value={dateFrom} onChange={e => setDateFrom(e.target.value)} className="w-full" />
                </div>
                <div>
                    <label className="block font-medium">Date To (optional)</label>
                    <input type="date" value={dateTo} onChange={e => setDateTo(e.target.value)} className="w-full" />
                </div>
            </div>
            <button onClick={handleGenerate} disabled={loading} className="btn-primary mt-4">Generate Report</button>

            {loading && <div>Generating report...</div>}
            {result && (
                <div className="mt-6">
                    <div className="flex justify-between items-center">
                        <h2 className="font-semibold">Result</h2>
                        {format === "csv" && <button onClick={downloadCSV} className="btn-secondary">Download CSV</button>}
                    </div>
                    <pre className="bg-gray-100 p-4 rounded overflow-auto text-sm mt-2">
                        {format === "json" ? JSON.stringify(result, null, 2) : result}
                    </pre>
                </div>
            )}
        </div>
    );
}