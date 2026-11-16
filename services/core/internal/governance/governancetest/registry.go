package governancetest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// PollLabel is the label every poll round carries in the реєстр (AC-41).
const PollLabel = "Попереднє внутрішнє опитування правління — не є голосуванням засідання і не замінює його"

// SignedSpec describes the one decision of a signed protocol that SeedSigned writes.
type SignedSpec struct {
	Number      int
	MeetingDate time.Time // a calendar date
	Subject     string
	// Counterparty is empty when the decision names none.
	Counterparty string
	// AmountKopiyky is nil when the decision has no amount; the currency is then null too.
	AmountKopiyky *int64
	// LockedBy is the board member who locked the protocol (board_members.id).
	LockedBy uuid.UUID
	// CorrectionsNote is the note of the handwritten corrections; empty means none were reported.
	CorrectionsNote string
	// PrintBytes and ScanBytes, when set, are the real content of the print copy and the signed scan: their
	// sha256 and length replace the synthetic ones in the files rows and the snapshot, so a download test
	// can serve bytes the stored digest really matches.
	PrintBytes, ScanBytes []byte
}

// Signed is what SeedSigned wrote, for the assertions of a test.
type Signed struct {
	ProtocolID, SnapshotID, MeetingID, AgendaItemID uuid.UUID
	PrintCopyID, PrintFileID, ScanFileID            uuid.UUID
	DocumentID, DocumentFileID, PollRoundID         uuid.UUID
	PrintFileName, ScanFileName, DocumentFileName   string
	PrintSHA256, ScanSHA256, DocumentSHA256         string // hex
	LockedAt                                        time.Time
	// DriveIDs is every Drive file and revision id written for this protocol. The snapshot and the files
	// rows hold them (ADR-0005, ADR-0007); no response body may ever contain one.
	DriveIDs []string
}

// SeedSigned writes a signed protocol the way the lock path of board-protocol T29 leaves it, on the
// transaction tx (db.InAssociation): the working rows, the signed snapshot with its fingerprint, the
// registry entry of its one decision, and a files row with Drive ids for the print copy, the signed
// scan and one cited case document. Nothing in the app can lock a protocol before T29, so the
// реєстр tests build their data here.
//
// The snapshot is schema_version 1. Its payload holds, under the names of the contract's SignedProtocol
// schema, meeting, agenda_items, documents, print_copy, meeting_facts, poll_history, locked_by_member_id
// and locked_at; every file object carries the FileRef fields plus drive_file_id and drive_revision_id.
// protocol_id, number and meeting_date are the columns of the snapshot row, not part of the payload.
func SeedSigned(t testing.TB, tx db.DBTX, association uuid.UUID, spec SignedSpec) Signed {
	t.Helper()
	ctx := context.Background()
	id := func() uuid.UUID { return uuid.Must(uuid.NewV7()) }
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("governancetest: %v\n%s", err, sql)
		}
	}
	s := Signed{
		ProtocolID: id(), SnapshotID: id(), MeetingID: id(), AgendaItemID: id(),
		PrintCopyID: id(), PrintFileID: id(), ScanFileID: id(),
		DocumentID: id(), DocumentFileID: id(), PollRoundID: id(),
		PrintFileName: "protocol-signing-copy.pdf", ScanFileName: "protocol-signed-scan.pdf", DocumentFileName: "invoice-cited.pdf",
		PrintSHA256:    sha(1),
		ScanSHA256:     sha(2),
		DocumentSHA256: sha(3),
		LockedAt:       spec.MeetingDate.Add(40 * time.Hour).UTC(),
	}
	printSize, scanSize := int64(91234), int64(1834221)
	if spec.PrintBytes != nil {
		sum := sha256.Sum256(spec.PrintBytes)
		s.PrintSHA256, printSize = hex.EncodeToString(sum[:]), int64(len(spec.PrintBytes))
	}
	if spec.ScanBytes != nil {
		sum := sha256.Sum256(spec.ScanBytes)
		s.ScanSHA256, scanSize = hex.EncodeToString(sum[:]), int64(len(spec.ScanBytes))
	}
	tag := s.ProtocolID.String()
	drive := map[string][2]string{
		"print": {"gdrv-file-print-" + tag, "gdrv-rev-print-" + tag},
		"scan":  {"gdrv-file-scan-" + tag, "gdrv-rev-scan-" + tag},
		"doc":   {"gdrv-file-doc-" + tag, "gdrv-rev-doc-" + tag},
	}
	for _, k := range []string{"print", "scan", "doc"} {
		s.DriveIDs = append(s.DriveIDs, drive[k][0], drive[k][1])
	}
	heldAt := spec.MeetingDate.Add(16 * time.Hour).UTC()
	caseID, round := id(), 1
	rfc := func(x time.Time) string { return x.UTC().Format(time.RFC3339) }

	exec(`INSERT INTO meetings (id, title, held_at, association_id, chair_board_member_id) VALUES ($1, 'Test meeting', $2, $3, $4)`,
		s.MeetingID, heldAt, association, spec.LockedBy)
	exec(`INSERT INTO protocols (id, association_id, meeting_id, state, revision, number, created_by_member_id)
		VALUES ($1, $2, $3, 'signed', 9, $4, $5)`, s.ProtocolID, association, s.MeetingID, spec.Number, spec.LockedBy)
	exec(`INSERT INTO cases (id, association_id, title, created_by_member_id) VALUES ($1, $2, 'Test case', $3)`, caseID, association, spec.LockedBy)
	file := func(fileID uuid.UUID, key, name, sum string, size int64, uploader any) {
		b, err := hex.DecodeString(sum)
		if err != nil {
			t.Fatal(err)
		}
		exec(`INSERT INTO files (id, association_id, drive_file_id, drive_revision_id, sha256, size_bytes, media_type, original_name, uploaded_by_member_id)
			VALUES ($1, $2, $3, $4, $5, $6, 'application/pdf', $7, $8)`,
			fileID, association, drive[key][0], drive[key][1], b, size, name, uploader)
	}
	file(s.PrintFileID, "print", s.PrintFileName, s.PrintSHA256, printSize, nil)
	file(s.ScanFileID, "scan", s.ScanFileName, s.ScanSHA256, scanSize, spec.LockedBy)
	file(s.DocumentFileID, "doc", s.DocumentFileName, s.DocumentSHA256, 482113, spec.LockedBy)
	exec(`INSERT INTO case_documents (id, association_id, case_id, file_id, created_by_member_id) VALUES ($1, $2, $3, $4, $5)`,
		s.DocumentID, association, caseID, s.DocumentFileID, spec.LockedBy)
	exec(`INSERT INTO agenda_items (id, association_id, protocol_id, case_id, position, created_by_member_id) VALUES ($1, $2, $3, $4, 1, $5)`,
		s.AgendaItemID, association, s.ProtocolID, caseID, spec.LockedBy)
	exec(`INSERT INTO print_copies (id, association_id, protocol_id, protocol_revision, number, file_id, requested_by_member_id)
		VALUES ($1, $2, $3, 8, $4, $5, $6)`, s.PrintCopyID, association, s.ProtocolID, spec.Number, s.PrintFileID, spec.LockedBy)
	factsID := id()
	var note any
	if spec.CorrectionsNote != "" {
		note = spec.CorrectionsNote
	}
	exec(`INSERT INTO meeting_facts (id, association_id, protocol_id, print_copy_id, signed_scan_file_id, has_handwritten_corrections, corrections_note, entered_by_member_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		factsID, association, s.ProtocolID, s.PrintCopyID, s.ScanFileID, spec.CorrectionsNote != "", note, spec.LockedBy)
	exec(`INSERT INTO meeting_attendees (meeting_facts_id, board_member_id, association_id) VALUES ($1, $2, $3)`, factsID, spec.LockedBy, association)
	exec(`INSERT INTO meeting_item_votes (meeting_facts_id, agenda_item_id, association_id, votes_for, votes_against, votes_abstained)
		VALUES ($1, $2, $3, 3, 0, 0)`, factsID, s.AgendaItemID, association)
	exec(`INSERT INTO poll_rounds (id, association_id, protocol_id, round_number, protocol_revision, frozen, opened_by_member_id)
		VALUES ($1, $2, $3, $4, 7, '[]'::jsonb, $5)`, s.PollRoundID, association, s.ProtocolID, round, spec.LockedBy)

	fileObj := func(fileID uuid.UUID, key, name, sum string, size int64, uploader any) map[string]any {
		return map[string]any{
			"id": fileID.String(), "original_name": name, "media_type": "application/pdf", "size_bytes": size, "sha256": sum,
			"uploaded_by_member_id": uploader, "created_at": rfc(s.LockedAt.Add(-24 * time.Hour)),
			"drive_file_id": drive[key][0], "drive_revision_id": drive[key][1],
		}
	}
	by := spec.LockedBy.String()
	payload := map[string]any{
		"meeting": map[string]any{
			"id": s.MeetingID.String(), "held_at": rfc(heldAt), "meeting_date": spec.MeetingDate.Format("2006-01-02"), "chair_board_member_id": by,
		},
		"agenda_items": []any{map[string]any{
			"id": s.AgendaItemID.String(), "position": 1, "case_id": caseID.String(), "subject": spec.Subject,
			"fragments":                []any{map[string]any{"text": spec.Subject}},
			"wording_author_member_id": by, "wording_revised_at": rfc(s.LockedAt.Add(-48 * time.Hour)), "slots": []any{},
		}},
		"documents": []any{map[string]any{
			"id": s.DocumentID.String(), "case_id": caseID.String(),
			"file":                fileObj(s.DocumentFileID, "doc", s.DocumentFileName, s.DocumentSHA256, 482113, by),
			"previous_version_id": nil, "document_date": nil, "created_by_member_id": by, "created_at": rfc(s.LockedAt.Add(-72 * time.Hour)),
		}},
		"print_copy": map[string]any{
			"id": s.PrintCopyID.String(), "protocol_revision": 8, "number": spec.Number,
			"file":                   fileObj(s.PrintFileID, "print", s.PrintFileName, s.PrintSHA256, printSize, nil),
			"requested_by_member_id": by, "created_at": rfc(s.LockedAt.Add(-30 * time.Hour)), "voided_at": nil,
		},
		"meeting_facts": map[string]any{
			"signed_scan":        fileObj(s.ScanFileID, "scan", s.ScanFileName, s.ScanSHA256, scanSize, by),
			"present_member_ids": []any{by},
			"item_votes": []any{map[string]any{
				"agenda_item_id": s.AgendaItemID.String(), "votes_for": 3, "votes_against": 0, "votes_abstained": 0,
			}},
			"has_handwritten_corrections": spec.CorrectionsNote != "",
			"corrections_note":            nilIfEmpty(spec.CorrectionsNote),
			"entered_by_member_id":        by, "entered_at": rfc(s.LockedAt),
		},
		"poll_history": []any{map[string]any{
			"id": s.PollRoundID.String(), "round_number": round, "protocol_revision": 7, "label": PollLabel,
			"opened_by_member_id": by, "created_at": rfc(s.LockedAt.Add(-60 * time.Hour)), "closed_at": nil, "closed_by_member_id": nil,
			"frozen_agenda_items": []any{},
			"items": []any{map[string]any{
				"agenda_item_id": s.AgendaItemID.String(),
				"tally":          map[string]any{"for": 1, "against": 0, "abstain": 0},
				"votes": []any{map[string]any{
					"board_member_id": by, "choice": "for", "created_at": rfc(s.LockedAt.Add(-59 * time.Hour)),
				}},
			}},
		}},
		"locked_by_member_id": by,
		"locked_at":           rfc(s.LockedAt),
	}
	snapshot := writeSnapshot(t, tx, association, s.SnapshotID, s.ProtocolID, domain.SnapshotSigned, spec.Number, spec.MeetingDate, spec.LockedBy, payload)

	var counterparty, currency any
	if spec.Counterparty != "" {
		counterparty = spec.Counterparty
	}
	var amount any
	if spec.AmountKopiyky != nil {
		amount, currency = *spec.AmountKopiyky, "UAH"
	}
	exec(`INSERT INTO registry_entries (id, association_id, snapshot_id, agenda_item_id, number, meeting_date, subject, counterparty, amount_kopiyky, currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		id(), association, snapshot, s.AgendaItemID, spec.Number, spec.MeetingDate, spec.Subject, counterparty, amount, currency)
	return s
}

// SeedWithdrawn writes a withdrawn protocol that holds number, with its withdrawn snapshot and no registry
// entry: the registry shows it only as a gap marked «знято з розгляду» (ADR-0007, ADR-0017).
func SeedWithdrawn(t testing.TB, tx db.DBTX, association uuid.UUID, number int, meetingDate time.Time, by uuid.UUID) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	meetingID, protocolID := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	if _, err := tx.Exec(ctx, `INSERT INTO meetings (id, title, held_at, association_id) VALUES ($1, 'Test meeting', $2, $3)`,
		meetingID, meetingDate.Add(16*time.Hour).UTC(), association); err != nil {
		t.Fatalf("governancetest: insert meeting: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO protocols (id, association_id, meeting_id, state, revision, number, created_by_member_id)
		VALUES ($1, $2, $3, 'withdrawn', 5, $4, $5)`, protocolID, association, meetingID, number, by); err != nil {
		t.Fatalf("governancetest: insert protocol: %v", err)
	}
	writeSnapshot(t, tx, association, uuid.Must(uuid.NewV7()), protocolID, domain.SnapshotWithdrawn, number, meetingDate, by,
		map[string]any{"reason": "Не підписано на засіданні", "withdrawn_by_member_id": by.String()})
	return protocolID
}

// writeSnapshot inserts the snapshot row with the fingerprint of its canonical form.
func writeSnapshot(t testing.TB, tx db.DBTX, association, snapshotID, protocolID uuid.UUID, kind domain.SnapshotKind,
	number int, meetingDate time.Time, by uuid.UUID, payload map[string]any) uuid.UUID {
	t.Helper()
	snap := domain.Snapshot{SchemaVersion: 1, Kind: kind, ProtocolID: protocolID.String(), Number: number, MeetingDate: meetingDate, Payload: payload}
	fp, err := snap.Fingerprint()
	if err != nil {
		t.Fatalf("governancetest: fingerprint: %v", err)
	}
	sum, err := hex.DecodeString(fp)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(context.Background(), `INSERT INTO protocol_snapshots
		(id, association_id, protocol_id, kind, number, meeting_date, schema_version, snapshot, fingerprint, created_by_member_id)
		VALUES ($1, $2, $3, $4, $5, $6, 1, $7::jsonb, $8, $9)`,
		snapshotID, association, protocolID, string(kind), number, meetingDate, string(doc), sum, by); err != nil {
		t.Fatalf("governancetest: insert snapshot: %v", err)
	}
	return snapshotID
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// sha is a distinct 64-hex digest per n.
func sha(n int) string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(n*31 + i)
	}
	return hex.EncodeToString(b)
}
