<?php

namespace App\Enums;

enum Role: string
{
    /** Manages events, payments and staff. */
    case Admin = 'admin';

    /** Checks people in at the door and reads attendee lists. */
    case Staff = 'staff';

    public function label(): string
    {
        return match ($this) {
            self::Admin => 'Administrador',
            self::Staff => 'Logística',
        };
    }
}
