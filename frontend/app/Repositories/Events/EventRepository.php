<?php

namespace App\Repositories\Events;

use App\Repositories\Contracts\Events\EventRepositoryInterface;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;

class EventRepository implements EventRepositoryInterface
{
    /**
     * Resolves an API endpoint from configuration with environment fallback
     */
    private function getEndpoint(string $key, string $envFallback): ?string
    {
        return config("services.events_api.{$key}", env($envFallback));
    }

    /**
     * Fetch all items by metadata parameter (e.g. 'EVENT' or 'USER')
     */
    public function getAll(string $name)
    {
        $url = $this->getEndpoint('show', 'URL_SHOW_EVENTS_USERS');
        if (empty($url)) {
            Log::warning('URL_SHOW_EVENTS_USERS endpoint is not configured');
            return [];
        }

        try {
            $response = Http::timeout(10)->get($url, ['param' => $name]);
            if ($response->successful()) {
                return $response->json() ?? [];
            }
            Log::error('Events microservice query failed', [
                'status' => $response->status(),
                'body' => $response->body()
            ]);
            return [];
        } catch (\Throwable $e) {
            Log::error('Events microservice connection error in getAll', ['error' => $e->getMessage()]);
            return [];
        }
    }

    /**
     * Persist an event or attendee entity to the serverless DynamoDB backend
     */
    public function createItem(array $data): bool
    {
        $url = $this->getEndpoint('create', 'URL_CREATE_EVENTS');
        if (empty($url)) {
            Log::warning('URL_CREATE_EVENTS endpoint is not configured');
            return false;
        }

        try {
            $response = Http::timeout(10)->post($url, $data);
            return $response->successful();
        } catch (\Throwable $e) {
            Log::error('Events microservice connection error in createItem', ['error' => $e->getMessage()]);
            return false;
        }
    }

    /**
     * Update an entity's lifecycle status
     */
    public function changeStatus($pk, $sk, $status): bool
    {
        $url = $this->getEndpoint('change_status', 'URL_CHANGE_STATUS_USERS_EVENTS');
        if (empty($url)) {
            Log::warning('URL_CHANGE_STATUS_USERS_EVENTS endpoint is not configured');
            return false;
        }

        try {
            $response = Http::timeout(10)->post($url, [
                'pk' => $pk,
                'sk' => $sk,
                'status' => $status
            ]);
            return $response->successful();
        } catch (\Throwable $e) {
            Log::error('Events microservice connection error in changeStatus', ['error' => $e->getMessage()]);
            return false;
        }
    }

    /**
     * Fetch attendees registered for an event
     */
    public function getAttendeesByEvent($eventID)
    {
        $url = $this->getEndpoint('event_users', 'URL_EVENT_USERS');
        if (empty($url)) {
            Log::warning('URL_EVENT_USERS endpoint is not configured');
            return [];
        }

        try {
            $response = Http::timeout(10)->get($url, ['param' => $eventID]);
            if ($response->successful()) {
                return $response->json() ?? [];
            }
            Log::error('Attendees microservice query failed', [
                'status' => $response->status(),
                'body' => $response->body()
            ]);
            return [];
        } catch (\Throwable $e) {
            Log::error('Attendees microservice connection error in getAttendeesByEvent', ['error' => $e->getMessage()]);
            return [];
        }
    }

    /**
     * Update attendee payment status (PAID/PENDING)
     */
    public function changePayStatus($pk, $sk, $status): bool
    {
        $url = $this->getEndpoint('change_pay_status', 'URL_CHANGE_PAY_STATUS');
        if (empty($url)) {
            Log::warning('URL_CHANGE_PAY_STATUS endpoint is not configured');
            return false;
        }

        try {
            $response = Http::timeout(10)->post($url, [
                'pk' => $pk,
                'sk' => $sk,
                'status' => $status,
                'paymentstatus' => $status
            ]);
            return $response->successful();
        } catch (\Throwable $e) {
            Log::error('Payment status microservice connection error in changePayStatus', ['error' => $e->getMessage()]);
            return false;
        }
    }
}
