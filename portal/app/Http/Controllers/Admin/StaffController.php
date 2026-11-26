<?php

namespace App\Http\Controllers\Admin;

use App\Enums\Role;
use App\Http\Controllers\Controller;
use App\Models\User;
use Illuminate\Contracts\View\View;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Illuminate\Validation\Rule;
use Illuminate\Validation\Rules\Password;

final class StaffController extends Controller
{
    public function index(): View
    {
        return view('admin.staff.index', ['users' => User::query()->orderBy('name')->get()]);
    }

    public function create(): View
    {
        return view('admin.staff.create', ['roles' => Role::cases()]);
    }

    public function store(Request $request): RedirectResponse
    {
        $data = $request->validate([
            'name' => ['required', 'string', 'max:100'],
            'email' => ['required', 'email:rfc', 'max:254', Rule::unique('users', 'email')],
            'role' => ['required', Rule::enum(Role::class)],
            'password' => ['required', 'confirmed', Password::min(12)->letters()->numbers()],
        ]);

        $user = User::query()->create($data);

        return redirect()->route('admin.staff.index')->with('status', "Cuenta de {$user->name} creada.");
    }
}
