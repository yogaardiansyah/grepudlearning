'use client';

import { useState, FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import api from '../../../lib/api'; // Sesuaikan path jika berbeda

export default function RegisterPage() {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const router = useRouter();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null); // Reset error sebelum submit baru
    setSuccess(null);

    try {
      const response = await api.post('/register', {
        username,
        email,
        password,
      });

      if (response.status === 201) {
        setSuccess('Registrasi berhasil! Cek email Anda untuk kode verifikasi.');
        // Opsional: Redirect setelah beberapa detik
          setTimeout(() => {
          // Redirect ke halaman Verify dan bawa emailnya
          router.push(`/auth/verify?email=${encodeURIComponent(email)}`);
        }, 1500); // Jeda sebentar agar user baca pesan sukses
      }
    } catch (err: any) {
      if (err.response) {
        // Ambil pesan error dari backend Go
        setError(err.response.data.error || 'Terjadi kesalahan saat registrasi.');
      } else {
        setError('Tidak dapat terhubung ke server.');
      }
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <div className="w-full max-w-md p-8 space-y-6 bg-white rounded-lg shadow-md">
        <h1 className="text-2xl font-bold text-center text-gray-800">Buat Akun Baru</h1>
        
        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label htmlFor="username" className="block text-sm font-medium text-gray-700">
              Username
            </label>
            <input
              id="username"
              name="username"
              type="text"
              required
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700">
              Email
            </label>
            <input
              id="email"
              name="email"
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700">
              Password
            </label>
            <input
              id="password"
              name="password"
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>
          <div>
            <button
              type="submit"
              className="w-full px-4 py-2 font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              Register
            </button>
          </div>
        </form>

        {error && (
          <div className="p-3 text-sm text-center text-red-700 bg-red-100 rounded-md" role="alert">
            {error}
          </div>
        )}
        {success && (
          <div className="p-3 text-sm text-center text-green-700 bg-green-100 rounded-md" role="alert">
            {success}
          </div>
        )}
      </div>
    </div>
  );
}