package broker

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

const maxSegmentSize = 10

// Segment represents a portion of a partition log.
// Each segment starts at a baseOffset and stores a contiguous sequence of messages.
type Segment struct {
	baseOffset int
	messages   []string
	file       *os.File
}

// Partition represents a topic partition composed of multiple log segments.
// Each segment stores a contiguous range of message offsets and persists to disk.
type Partition struct {
	topic    string
	id       int
	segments []*Segment
	mu       sync.RWMutex
}

// NewPartition creates a new partition and loads existing segments from disk.
func NewPartition(topic string, id int) (*Partition, error) {
	os.MkdirAll("storage", os.ModePerm)

	p := &Partition{
		topic:    topic,
		id:       id,
		segments: []*Segment{},
	}

	// Load existing segments from disk
	if err := p.loadSegmentsFromDisk(); err != nil {
		return nil, err
	}

	return p, nil
}

// loadSegmentsFromDisk reads all existing log files for this partition from disk.
func (p *Partition) loadSegmentsFromDisk() error {
	files, err := os.ReadDir("storage")
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("%s-%d-", p.topic, p.id)

	for _, entry := range files {
		name := entry.Name()

		if !strings.HasPrefix(name, prefix) {
			continue
		}

		var baseOffset int
		_, err := fmt.Sscanf(name, prefix+"%d.log", &baseOffset)
		if err != nil {
			continue
		}

		filename := fmt.Sprintf("storage/%s", name)

		f, err := os.OpenFile(
			filename,
			os.O_APPEND|os.O_CREATE|os.O_RDWR,
			0644,
		)
		if err != nil {
			return err
		}

		seg := &Segment{
			baseOffset: baseOffset,
			messages:   []string{},
			file:       f,
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			seg.messages = append(seg.messages, line)
		}
		if err := scanner.Err(); err != nil {
			return err
		}

		p.segments = append(p.segments, seg)
	}

	return nil
}

// Append writes a message to the partition, rotating segments when needed.
// Returns the offset of the new message or an error if the write operation fails.
func (p *Partition) Append(message string) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var seg *Segment

	// Create first segment if none exist
	if len(p.segments) == 0 {
		s, err := p.newSegment(0)
		if err != nil {
			return 0, err
		}
		seg = s
	} else {
		seg = p.segments[len(p.segments)-1]
	}

	offset := seg.baseOffset + len(seg.messages)

	// Create new segment if current one is full
	if len(seg.messages) >= maxSegmentSize {
		s, err := p.newSegment(offset)
		if err != nil {
			return 0, err
		}
		seg = s
	}

	// Append to memory and disk
	seg.messages = append(seg.messages, message)
	_, err := seg.file.WriteString(message + "\n")
	if err != nil {
		return 0, err
	}

	return offset, nil
}

// newSegment creates a new segment starting at baseOffset.
func (p *Partition) newSegment(baseOffset int) (*Segment, error) {
	filename := fmt.Sprintf("storage/%s-%d-%d.log", p.topic, p.id, baseOffset)

	file, err := os.OpenFile(
		filename,
		os.O_APPEND|os.O_CREATE|os.O_RDWR,
		0644,
	)
	if err != nil {
		return nil, err
	}

	s := &Segment{
		baseOffset: baseOffset,
		messages:   []string{},
		file:       file,
	}

	p.segments = append(p.segments, s)

	return s, nil
}

// Fetch retrieves messages from the partition starting at the given offset.
func (p *Partition) Fetch(offset int64) []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var messages []string

	for _, seg := range p.segments {
		if int(offset) >= seg.baseOffset+len(seg.messages) {
			continue
		}
		for i, msg := range seg.messages {
			msgOffset := seg.baseOffset + i
			if msgOffset >= int(offset) {
				messages = append(messages, fmt.Sprintf("offset %d: %s", msgOffset, msg))
			}
		}
	}

	return messages
}
