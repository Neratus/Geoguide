import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider, useAuth } from "./contexts/AuthContext";
import Layout from "./components/Layout/Layout";
import LoginPage from "./pages/LoginPage";
import RegisterPage from "./pages/RegisterPage";
import CountriesPage from "./pages/CountriesPage";
import CountryDetailPage from "./pages/CountryDetailPage";
import CityDetailPage from "./pages/CityDetailPage";
import PlaceDetailPage from "./pages/PlaceDetailPage";
import ProfilePage from "./pages/ProfilePage";
import TripsPage from "./pages/TripsPage";
import CreateTripPage from "./pages/CreateTripPage";
import TripDetailPage from "./pages/TripDetailPage";
import ReviewsPage from "./pages/ReviewsPage";
import CreateReviewPage from "./pages/CreateReviewPage";
import ModerateReviewsPage from "./pages/ModerateReviewsPage";
import AnalyticsPage from "./pages/AnalyticsPage";
import TwoFactorPage from "./pages/TwoFactorPage";
import VerifyPage from "./pages/VerifyPage";
import ModerateUsersPage from "./pages/ModerateUsersPage";
import FavouritesPage from "./pages/FavouritesPage";

const RoleRoute = ({ children, allowedRoles }: { children: JSX.Element; allowedRoles: string[] }) => {
  const { user, token, isLoading } = useAuth();
  if (isLoading) return <div className="p-8 text-center">Loading...</div>;
  if (!token) return <Navigate to="/login" replace />;
  if (!user || !allowedRoles.includes(user.role)) return <Navigate to="/" replace />;
  return children;
};

const ProtectedRoute = ({ children }: { children: JSX.Element }) => {
  const { token, isLoading } = useAuth();
  if (isLoading) return <div className="p-8 text-center">Loading...</div>;
  if (!token) return <Navigate to="/login" replace />;
  return children;
};

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/" element={<Layout />}>
            <Route path="countries" element={<CountriesPage />} />
            <Route path="countries/:id" element={<CountryDetailPage />} />
            <Route path="cities/:id" element={<CityDetailPage />} />
            <Route path="places/:id" element={<PlaceDetailPage />} />
            <Route path="profile" element={<ProtectedRoute><ProfilePage /></ProtectedRoute>} />
            <Route path="trips" element={<ProtectedRoute><TripsPage /></ProtectedRoute>} />
            <Route path="trips/new" element={<ProtectedRoute><CreateTripPage /></ProtectedRoute>} />
            <Route path="trips/:id" element={<ProtectedRoute><TripDetailPage /></ProtectedRoute>} />
            <Route path="reviews" element={<ProtectedRoute><ReviewsPage /></ProtectedRoute>} />
            <Route path="places/:placeId/reviews/new" element={<ProtectedRoute><CreateReviewPage /></ProtectedRoute>} />
            <Route path="moderate/reviews" element={<RoleRoute allowedRoles={["MODERATOR", "ANALYST"]}><ModerateReviewsPage /></RoleRoute>} />
            <Route path="analytics" element={<RoleRoute allowedRoles={["ANALYST"]}><AnalyticsPage /></RoleRoute>} />
            <Route path="/moderate/reviews" element={<RoleRoute allowedRoles={["MODERATOR", "ANALYST"]}><ModerateReviewsPage /></RoleRoute>} />
            <Route path="/moderate/users" element={<RoleRoute allowedRoles={["MODERATOR"]}><ModerateUsersPage /></RoleRoute>} />
            <Route path="favourites" element={<ProtectedRoute><FavouritesPage /></ProtectedRoute>} />
          </Route>
          <Route path="/2fa" element={<TwoFactorPage />} />
          <Route path="/verify" element={<VerifyPage />} />
          <Route path="/2fa/verify" element={<TwoFactorPage />} />
          <Route path="/verify" element={<VerifyPage />} />

        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
export default App;