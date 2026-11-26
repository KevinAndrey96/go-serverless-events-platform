<?php

namespace Tests\Feature;

use App\Enums\Role;
use App\Models\User;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\Hash;
use Tests\TestCase;

final class AuthAndStaffTest extends TestCase
{
    use RefreshDatabase;

    public function test_staff_sign_in_and_out(): void
    {
        $user = User::factory()->create(['email' => 'equipo@example.com']);

        $this->post(route('login.store'), ['email' => 'equipo@example.com', 'password' => 'password'])
            ->assertRedirect(route('admin.events.index'));
        $this->assertAuthenticatedAs($user);

        $this->post(route('admin.logout'))->assertRedirect(route('login'));
        $this->assertGuest();
    }

    public function test_wrong_passwords_are_rejected_and_throttled(): void
    {
        User::factory()->create(['email' => 'equipo@example.com']);

        foreach (range(1, 5) as $attempt) {
            $this->post(route('login.store'), ['email' => 'equipo@example.com', 'password' => 'incorrecta'])
                ->assertSessionHasErrors(['email' => 'El correo o la contraseña no son correctos.']);
        }

        $this->post(route('login.store'), ['email' => 'equipo@example.com', 'password' => 'password'])
            ->assertSessionHasErrors('email');
        $this->assertStringContainsString('Demasiados intentos', session('errors')->first('email'));
        $this->assertGuest();
    }

    public function test_signed_in_users_skip_the_login(): void
    {
        $this->actingAs(User::factory()->create())->get(route('login'))->assertRedirect(route('admin.events.index'));
    }

    public function test_admins_create_staff_accounts(): void
    {
        $this->actingAs(User::factory()->admin()->create())
            ->post(route('admin.staff.store'), [
                'name' => 'Logística Puerta 1',
                'email' => 'puerta1@example.com',
                'role' => 'staff',
                'password' => 'clave-segura-2026',
                'password_confirmation' => 'clave-segura-2026',
            ])
            ->assertRedirect(route('admin.staff.index'));

        $user = User::query()->where('email', 'puerta1@example.com')->sole();
        $this->assertSame(Role::Staff, $user->role);
        $this->assertTrue(Hash::check('clave-segura-2026', $user->password));
    }

    public function test_staff_accounts_need_strong_unique_credentials(): void
    {
        User::factory()->create(['email' => 'puerta1@example.com']);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('admin.staff.store'), [
                'name' => 'Otra',
                'email' => 'puerta1@example.com',
                'role' => 'root',
                'password' => 'corta',
                'password_confirmation' => 'distinta',
            ])
            ->assertSessionHasErrors(['email', 'role', 'password']);
    }

    public function test_seeder_creates_the_admin_only_with_a_configured_password(): void
    {
        config(['events.admin.email' => 'admin@example.com', 'events.admin.password' => 'corta']);
        $this->seed();
        $this->assertDatabaseMissing('users', ['email' => 'admin@example.com']);

        config(['events.admin.password' => 'una-clave-larga-2026']);
        $this->seed();
        $this->assertSame(Role::Admin, User::query()->where('email', 'admin@example.com')->sole()->role);
    }
}
