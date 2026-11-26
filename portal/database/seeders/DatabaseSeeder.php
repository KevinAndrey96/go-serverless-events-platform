<?php

namespace Database\Seeders;

use App\Enums\Role;
use App\Models\User;
use Illuminate\Database\Seeder;

class DatabaseSeeder extends Seeder
{
    /**
     * Creates the first administrator from ADMIN_EMAIL and ADMIN_PASSWORD; nothing is hardcoded.
     */
    public function run(): void
    {
        $email = config('events.admin.email');
        $password = config('events.admin.password');

        if (! is_string($email) || $email === '' || ! is_string($password) || strlen($password) < 12) {
            $this->command?->warn('ADMIN_EMAIL y ADMIN_PASSWORD (mínimo 12 caracteres) no están definidos; no se creó el administrador.');

            return;
        }

        User::query()->updateOrCreate(
            ['email' => $email],
            ['name' => 'Administrador', 'password' => $password, 'role' => Role::Admin],
        );
    }
}
