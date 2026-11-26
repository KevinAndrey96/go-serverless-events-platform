<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Model;

/**
 * Payment receipt uploaded at registration. Files live on the private disk and are only
 * served to signed-in staff.
 *
 * @property int $id
 * @property string $event_id
 * @property string $attendee_id
 * @property string $path
 * @property string $mime_type
 */
#[Fillable(['event_id', 'attendee_id', 'path', 'mime_type'])]
class Receipt extends Model {}
