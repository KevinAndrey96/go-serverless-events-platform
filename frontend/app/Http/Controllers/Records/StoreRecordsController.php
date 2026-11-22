<?php

namespace App\Http\Controllers\Records;

use App\Http\Controllers\Controller;
use App\Models\AttendeeFile;
use App\Repositories\Contracts\Events\EventRepositoryInterface;
use Barryvdh\DomPDF\Facade\Pdf;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Mail;
use Illuminate\Support\Facades\Storage;
use SimpleSoftwareIO\QrCode\Facades\QrCode;
use Webpatser\Uuid\Uuid;

class StoreRecordsController extends Controller
{
    private EventRepositoryInterface $eventRepository;

    public function __construct(EventRepositoryInterface $eventRepository)
    {
        $this->eventRepository = $eventRepository;
    }

    public function __invoke(Request $request)
    {
        $request->validate([
            'event_pk' => 'required',
            'first_name' => 'required|string',
            'last_name' => 'required|string',
            'email' => 'required|email',
        ]);

        $uuID = substr(''.Uuid::generate(), 0, 6);
        $data = [
            "pk" => "User#".$uuID,
            "sk" => "METADATA#USER",
            "eventID" => 'Event#'.strval($request->input('event_pk')),
            "name" => strval($request->input('first_name')),
            "lastname" => strval($request->input('last_name')),
            "gender" => strval($request->input('gender')),
            "email" => strval($request->input('email')),
            "phone" => strval($request->input('phone')),
            "date" => strval($request->input('date')),
            "status" => 'not-attend',
            "userID" => $uuID,
            "scholarship" => strval($request->input('scholarship')),
            "paymentstatus" => 'not-payed',
        ];

        // Register attendee in Go serverless microservice
        $this->eventRepository->createItem($data);

        $subject = 'Código de acceso para evento';
        $text = 'Buenos días.<br>Con este correo se ha adjuntado un código QR que le permitirá acceder al evento seleccionado.<br>Gracias por usar nuestros servicios.';

        $pk = $data['pk'];
        $qrCode = QrCode::format('png')->size(300)->generate($pk);
        $qrCode = base64_encode($qrCode);
        $pdf = Pdf::loadview('users.documentQR', compact('qrCode'));

        if ($request->hasFile('receiptFile')) {
            $pathName = sprintf('receipt_files/%s.png', $data['userID']);
            Storage::disk('public')->put($pathName, file_get_contents($request->file('receiptFile')));

            $attendee = new AttendeeFile();
            $attendee->url_receipt = '/storage/receipt_files/'.$data['userID'].'.png';
            $attendee->save();
        }

        try {
            Mail::send('users.mailQR', compact('text'), function($mail) use ($pdf, $data, $subject) {
                $mail->to($data['email']);
                $mail->subject($subject);
                $mail->attachData($pdf->output(), 'QR.pdf');
            });

            if ($request->filled('friend_email')) {
                $friendEmail = $request->input('friend_email');
                $inviteLink = url('/records-create/'.strval($request->input('event_pk')));
                $inviteText = 'Un amigo te ha referenciado para que puedas asistir a un evento. Puedes registrarte dando clic en el siguiente enlace: '.$inviteLink;
                $inviteSubject = 'Invitación a evento';

                Mail::send('users.mailQR', ['text' => $inviteText], function($mail) use ($inviteSubject, $friendEmail) {
                    $mail->to($friendEmail);
                    $mail->subject($inviteSubject);
                });
            }
        } catch (\Throwable $e) {
            \Illuminate\Support\Facades\Log::warning('Could not send notification email: '.$e->getMessage());
        }

        return redirect()->route('records.events')->with('recordRegistered', 'Ha sido registrado al evento con éxito');
    }
}
