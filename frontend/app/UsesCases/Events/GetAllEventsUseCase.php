<?php

namespace App\UsesCases\Events;

use App\Repositories\Contracts\Events\EventRepositoryInterface;
use App\UsesCases\Contracts\Events\GetAllEventsUseCaseInterface;

class GetAllEventsUseCase implements GetAllEventsUseCaseInterface
{
    private EventRepositoryInterface $eventRepository;

    public function __construct(EventRepositoryInterface $eventRepository)
    {
        $this->eventRepository = $eventRepository;
    }

    public function handle(string $name)
    {
        return $this->eventRepository->getAll($name);
    }
}
