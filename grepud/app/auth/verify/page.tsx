// app/auth/verify/page.tsx
'use client';

import { useState, FormEvent, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import api from '../../../lib/api'; 

function VerifyContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // Otomatis isi email jika ada di URL (contoh: /auth/verify?email=budi@mail.com)
  useEffect(() => {
    const emailParam = searchParams.get('email');
    if (emailParam) {
      setEmail(emailParam);
    }
  }, [searchParams]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    setLoading(true);

    try {
      // Panggil endpoint backend: POST /auth/verify
      const response = await api.post('/verify', {
        email,
        code, // Kode OTP 6 digit
      });

      if (response.status === 200) {
        setSuccess('Verifikasi berhasil! Mengalihkan ke halaman login...');
        setTimeout(() => {
          router.push('/auth/login');
        }, 2000);
      }
    } catch (err: any) {
      if (err.response) {
        setError(err.response.data.error || 'Kode salah atau kadaluarsa.');
      } else {
        setError('Gagal terhubung ke server.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <div className="w-full max-w-md p-8 space-y-6 bg-white rounded-lg shadow-md">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-800">Verifikasi Akun</h1>
          <p className="mt-2 text-sm text-gray-600">
            Masukkan 6 digit kode yang telah dikirim ke email: <br/>
            <span className="font-semibold text-indigo-600">{email}</span>
          </p>
        </div>
        
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Input Email (Bisa diedit jika user salah ketik saat register) */}
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700">
              Email
            </label>
            <input
              id="email"
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-3 py-2 mt-1 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>

          {/* Input Kode OTP */}
          <div>
            <label htmlFor="code" className="block text-sm font-medium text-gray-700">
              Kode Verifikasi (OTP)
            </label>
            <input
              id="code"
              type="text"
              required
              maxLength={6}
              placeholder="Contoh: 555181"
              value={code}
              onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))} // Hanya angka
              className="w-full px-3 py-2 mt-1 text-center text-lg tracking-widest border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>

          <div>
            <button
              type="submit"
              disabled={loading}
              className={`w-full px-4 py-2 font-medium text-white rounded-md focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 ${
                loading ? 'bg-indigo-400 cursor-not-allowed' : 'bg-indigo-600 hover:bg-indigo-700'
              }`}
            >
              {loading ? 'Memproses...' : 'Verifikasi Akun'}
            </button>
          </div>
        </form>

        {error && (
          <div className="p-3 text-sm text-center text-red-700 bg-red-100 rounded-md">
            {error}
          </div>
        )}
        {success && (
          <div className="p-3 text-sm text-center text-green-700 bg-green-100 rounded-md">
            {success}
          </div>
        )}
      </div>
    </div>
  );
}

// Wajib menggunakan Suspense saat menggunakan useSearchParams di Next.js App Router
export default function VerifyPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <VerifyContent />
    </Suspense>
  );
}